{-# LANGUAGE OverloadedStrings #-}

module Model where

import Data.Aeson (Value, withObject, withArray)
import Data.Aeson.Types (Value(..), (.:), parseEither, Parser)
import Data.Aeson.Key (fromText)
import qualified Data.Vector as V
import Data.Text (Text, pack, unpack, isSuffixOf)
import Text.Read (readMaybe)
import Data.Maybe (mapMaybe)
import Data.Scientific (toRealFloat)
import Data.Time.Clock.POSIX (posixSecondsToUTCTime)
import Data.Time.Clock (utctDay)
import Data.Time (Day)
import Network.HTTP.Req ((/:), (=:), req, https, runReq, defaultHttpConfig, jsonResponse, responseBody)
import qualified Network.HTTP.Req as Req
import qualified Data.Map as Map

import Types
    ( StatResult(..),
      WeatherAnomaly(WeatherAnomaly),
      StatDB,
      Moon(..),
      Forecast(..),
      ForecastElement(..),
      Wind(..),
      Metrics(..),
      Weather(..),
      City(..) )
import Statistics (isKeyInvalid, mean, stdDev, median, mode, detectAnomalies)

extractField :: Text -> [Value] -> Parser Text
extractField field (x:_) = withObject "weather[0]" (.: fromText field) x
extractField _ _ = fail "Weather array is empty"

roundValue :: Int -> Double -> Double
roundValue n x = fromIntegral (round (x * t) :: Int) / t
    where t = 10^n

getEmoji :: Text -> Bool -> Text
getEmoji condition' isNight =
    case condition' of
        "Thunderstorm" -> "⛈️"
        "Drizzle"      -> "🌦️"
        "Rain"         -> "🌧️"
        "Snow"         -> "☃️"
        "Mist"         -> "💭"
        "Smoke"        -> "💭"
        "Haze"         -> "💭"
        "Dust"         -> "💭"
        "Fog"          -> "💭"
        "Sand"         -> "💭"
        "Ash"          -> "💭"
        "Squall"       -> "💭"
        "Tornado"      -> "🌪️"
        "Clear"        -> if isNight then "🌙" else "☀️"
        "Clouds"       -> "☁️"
        "SunWithCloud" -> "🌤️"
        "CloudWithSun" -> "🌥️"
        _              -> "❓"

getCardinalDir :: Double -> (Text, Text)
getCardinalDir windDeg =
    -- Each cardinal direction represents a segment of 22.5 degrees
    let cardinalDirections =
            [ ("N", "⬇️")   -- 0/360 DEG
            , ("NNE", "↙️") -- 22.5 DEG
            , ("NE",  "↙️") -- 45 DEG
            , ("ENE", "↙️") -- 67.5 DEG
            , ("E",   "⬅️") -- 90 DEG
            , ("ESE", "↖️") -- 112.5 DEG
            , ("SE",  "↖️") -- 135 DEG
            , ("SSE", "↖️") -- 157.5 DEG
            , ("S",   "⬆️") -- 180 DEG
            , ("SSW", "↗️") -- 202.5 DEG
            , ("SW",  "↗️") -- 225 DEG
            , ("WSW", "↗️") -- 247.5 DEG
            , ("W",   "➡️") -- 270 DEG
            , ("WNW", "↘️") -- 292.5 DEG
            , ("NW",  "↘️") -- 315 DEG
            , ("NNW", "↘️") -- 337.5 DEG
            ]
        -- Computes "idx ≡ round(wind_deg / 22.5) (mod 16)"
        -- to ensure that values above 360 degrees or below 0 degrees
        -- "stay" bounded to the map
        idx = round (windDeg / 22.5) `mod` 16
    in cardinalDirections !! idx

getCityCoords :: Text -> Text -> IO (Either Text City)
getCityCoords city appid = runReq defaultHttpConfig $ do
    -- Fetch city coordinates
    let reqUri = https "api.openweathermap.org" /: "geo" /: "1.0" /: "direct"
    response <- req Req.GET reqUri Req.NoReqBody jsonResponse $
        "q" =: city <>
        "limit" =: (5 :: Int) <>
        "appid" =: appid

    -- Parse JSON response
    let respBody = responseBody response :: Value
    case respBody of
        Array arr | V.null arr -> pure $ Left "Cannot find this city"
        Array arr -> do
            let root = V.head arr
            case parseEither parseCoords root of
                Right coords -> pure $ Right $ City
                    { name = city
                    , lat = fst coords
                    , lon = snd coords
                    }
                Left err -> pure $ Left $ "JSON parsing error: " <> pack err
        _ -> pure $ Left "Unexpected response format"
    where
        parseCoords :: Value -> Parser (Double, Double)
        parseCoords = withObject "root" $ \obj -> do
            latitude <- obj .: "lat"
            longitude <- obj .: "lon"
            pure (toRealFloat latitude, toRealFloat longitude)

getCityWeather :: City -> Text -> IO (Either Text Weather)
getCityWeather city appid = runReq defaultHttpConfig $ do
    -- Fetch weather data
    let reqUri = https "api.openweathermap.org" /: "data" /: "3.0" /: "onecall"
    response <- req Req.GET reqUri Req.NoReqBody jsonResponse $
        "lat" =: lat city <>
        "lon" =: lon city <>
        "appid" =: appid <>
        "units" =: ("metric" :: Text) <>
        "exclude" =: ("minutely,hourly,daily,alerts" :: Text)

    -- Parse JSON response
    let resBody = responseBody response :: Value
    case parseWeather resBody of
        Right weather -> pure $ Right weather
        Left err -> pure $ Left $ "Unable to parse API request: " <> pack err
    where
        parseWeather :: Value -> Either String Weather
        parseWeather = parseEither (withObject "root" $ \root -> do
            -- Extract keys from JSON response
            current <- root .: "current"
            weatherArray <- current .: "weather"
            condTitle <- withArray "weather" (extractField "main" . V.toList) weatherArray
            condDesc <- withArray "weather" (extractField "description" . V.toList) weatherArray
            temp <- current .: "temp" :: Parser Double
            unixTs <- current .: "dt"
            icon <- withArray "weather" (extractField "icon" . V.toList) weatherArray
            feelsLikeTemp <- current .: "feels_like" :: Parser Double

            -- Format UNIX timestamp as 'YYYY-MM-DD'
            let utcTime = posixSecondsToUTCTime (fromIntegral (unixTs :: Int))
            let weatherDate = utctDay utcTime

            -- Set condition accordingly to weather description
            let conditionVal = case condDesc of
                    "few clouds"    -> "SunWithCloud"
                    "broken clouds" -> "CloudWithSun"
                    _               -> condTitle

            -- Get emoji from weather condition
            let isNight = "n" `isSuffixOf` icon
            let emojiVal = getEmoji conditionVal isNight

            pure $ Weather
                { date = weatherDate
                , temperature = (pack . show) temp
                , feelsLike = (pack . show) feelsLikeTemp
                , condition = condTitle
                , condEmoji = emojiVal 
                })

getCityMetrics :: City -> Text -> IO (Either Text Metrics)
getCityMetrics city appid = runReq defaultHttpConfig $ do
    -- Fetch weather data
    let reqUri = https "api.openweathermap.org" /: "data" /: "3.0" /: "onecall"
    response <- req Req.GET reqUri Req.NoReqBody jsonResponse $
        "lat" =: lat city <>
        "lon" =: lon city <>
        "appid" =: appid <>
        "units" =: ("metric" :: Text) <>
        "exclude" =: ("minutely,hourly,daily,alerts" :: Text)

    -- Parse JSON response
    let resBody = responseBody response :: Value
    case parseMetrics resBody of
        Right metrics -> pure $ Right metrics
        Left err -> pure $ Left $ "Unable to parse API request: " <> pack err
    where
        parseMetrics :: Value -> Either String Metrics
        parseMetrics = parseEither (withObject "root" $ \root -> do
            -- Extract keys from JSON response
            current <- root .: "current"
            pressure' <- current .: "pressure"
            humidity' <- current .: "humidity"
            dewPointVal <- current .: "dew_point" :: Parser Double
            uvi <- current .: "uvi"
            vs <- current .: "visibility"

            -- Round UV index
            let uv = round (uvi :: Double) :: Int

            pure $ Metrics { humidity = (pack . show) (humidity' :: Int)
                           , pressure = (pack . show) (pressure' :: Int)
                           , dewPoint = (pack . show) dewPointVal
                           , uvIndex = uv
                           , visibility = (pack . show) (round (vs / 1000 :: Double) :: Int)
                           }
            )

getCityWind :: City -> Text -> IO (Either Text Wind)
getCityWind city apiKey = runReq defaultHttpConfig $ do
    -- Fetch weather data
    let reqUri = https "api.openweathermap.org" /: "data" /: "3.0" /: "onecall"
    response <- req Req.GET reqUri Req.NoReqBody jsonResponse $
        "lat" =: lat city <>
        "lon" =: lon city <>
        "appid" =: apiKey <>
        "units" =: ("metric" :: Text) <>
        "exclude" =: ("minutely,hourly,daily,alerts" :: Text)

    -- Parse JSON response
    let resBody = responseBody response :: Value
    case parseWind resBody of
        Right wind -> pure $ Right wind
        Left err -> pure $ Left $ "Unable to parse API request: " <> pack err
    where
        parseWind :: Value -> Either String Wind
        parseWind = parseEither (withObject "root" $ \root -> do
            -- Extract keys from JSON response
            current <- root .: "current"
            windSpeed <- current .: "wind_speed" :: Parser Double
            windDegree <- current .: "wind_deg" :: Parser Double

            -- Get cardinal direction and direction arrow
            let (windDirection, windArrow) = getCardinalDir windDegree

            pure $ Wind { speed = (pack . show) windSpeed
                        , direction = windDirection
                        , arrow = windArrow
                        }
            )

getCityForecast :: City -> Text -> IO (Either Text Forecast)
getCityForecast city apiKey = runReq defaultHttpConfig $ do
    -- Fetch weather data
    let reqUri = https "api.openweathermap.org" /: "data" /: "3.0" /: "onecall"
    response <- req Req.GET reqUri Req.NoReqBody jsonResponse $
        "lat" =: lat city <>
        "lon" =: lon city <>
        "appid" =: apiKey <>
        "units" =: ("metric" :: Text) <>
        "exclude" =: ("current,minutely,hourly,alerts" :: Text)

    -- Parse JSON response
    let resBody = responseBody response :: Value
    case parseForecast resBody of
        Right fc -> pure $ Right fc
        Left err -> pure $ Left $ "Unable to parse API request: " <> pack err
    where
        parseForecast :: Value -> Either String Forecast
        parseForecast = parseEither (withObject "root" $ \root -> do
            -- Extract the daily array from the JSON object
            daily <- root .: "daily"

            -- Parse each element of the array into a ForecastElement object.
            -- We drop the first element because it represent the current day
            -- which is not so useful in this context
            fc <- withArray "daily" (mapM getFCFields . take 5 . drop 1 . V.toList) daily
            
            pure $ Forecast { forecast = fc })
        getFCFields :: Value ->  Parser ForecastElement
        getFCFields = withObject "daily element" $ \obj -> do
            -- Extract temperature
            tempMin <- obj .: "temp" >>= (.: "min") :: Parser Double
            tempMax <- obj .: "temp" >>= (.: "max") :: Parser Double
            weatherArray <- obj .: "weather"
            -- Get forecast date(UNIX timestamp)
            unixTS <- obj .: "dt"
            -- Get conditions and icon
            conditionVal <- withArray "weather" (extractField "main" . V.toList) weatherArray
            icon <- withArray "weather" (extractField "icon" . V.toList) weatherArray
            fl <- obj .: "feels_like" >>= (.: "day") :: Parser Double
            -- Get wind speed, gust and direction
            windSpeed <- obj .: "wind_speed" :: Parser Double
            windDegree <- obj .: "wind_deg" :: Parser Double

            -- Format UNIX timestamp as UTC time
            let utcTime = posixSecondsToUTCTime (fromIntegral (unixTS :: Int))
            let weatherDate = utctDay utcTime

            -- Get emoji from weather condition
            let isNight = "n" `isSuffixOf` icon
            let emojiVal = getEmoji conditionVal isNight

            -- Get cardinal direction and direction arrow
            let (windDirection, windArrow) = getCardinalDir windDegree

            pure $ ForecastElement
                { fcDate = weatherDate
                , fcMin = (pack . show) tempMin
                , fcMax = (pack . show) tempMax
                , fcCond = conditionVal
                , fcEmoji = emojiVal
                , fcFL = (pack . show) fl
                , fcWindSpeed = (pack . show) windSpeed
                , fcWindDir = windDirection
                , fcWindArrow = windArrow
                }

getMoon :: Text -> IO (Either Text Moon)
getMoon apiKey = runReq defaultHttpConfig $ do
    -- Fetch weather data
    let reqUri = https "api.openweathermap.org" /: "data" /: "3.0" /: "onecall"
    response <- req Req.GET reqUri Req.NoReqBody jsonResponse $
        "lat" =: (41.8933203 :: Double) <> -- Rome latitude
        "lon" =: (12.4829321 :: Double) <> -- Rome longitude
        "appid" =: apiKey <>
        "units" =: ("metric" :: Text) <>
        "exclude" =: ("minutely,hourly,current,alerts" :: Text)

    -- Parse JSON response
    let resBody = responseBody response :: Value
    case parseMoon resBody of
        Right moon -> pure $ Right moon
        Left err -> pure $ Left $ "Unable to parse API request: " <> pack err
    where
        parseMoon :: Value -> Either String Moon
        parseMoon = parseEither (withObject "root" $ \root -> do
            -- Extract keys from JSON response
            daily <- root .: "daily"
            moonValue <- withArray "daily" (\arr -> do
                if V.null arr
                    then fail "daily array is empty"
                    else withObject "daily element" (.: "moon_phase") (V.head arr)) daily

            -- Map moon phase to emoji and phase description
            let (icon, phase) = getMoonPhase moonValue
            
            -- Approximate moon illumination percentage using moon phase
            let moonPercentage = getMoonPercentage moonValue

            pure $ Moon { moonEmoji = icon
                        , moonPhase = phase
                        , moonProgress = moonPercentage
                        }
            )

        {- 0 and 1 are 'new moon',
        0.25 is 'first quarter moon',
        0.5 is 'full moon' and 0.75 is 'last quarter moon'.
        The periods in between are called 'waxing crescent',
        'waxing gibbous', 'waning gibbous' and 'waning crescent', respectively. -}
        getMoonPhase :: Double -> (Text, Text)
        getMoonPhase moonValue
            | moonValue == 0 || moonValue == 1 = ("🌑", "New Moon")
            | moonValue > 0 && moonValue < 0.25 = ("🌒", "Waxing Crescent")
            | moonValue == 0.25 = ("🌓", "First Quarter")
            | moonValue > 0.25 && moonValue < 0.5 = ("🌔", "Waxing Gibbous")
            | moonValue == 0.5 = ("🌕", "Full Moon")
            | moonValue > 0.5 && moonValue < 0.75 = ("🌖", "Waning Gibbous")
            | moonValue == 0.75 = ("🌗", "Last Quarter")
            | moonValue > 0.75 && moonValue < 1 = ("🌘", "Waning Crescent")
            | otherwise = ("❓", "Unknown moon phase")

        -- Convert OpenWeatherMap moon value to percentage using
        -- sin(\pi * moon_value)^2
        getMoonPercentage :: Double -> Text
        getMoonPercentage val =
            let percentage = round ((sin (pi * val) ** 2) * 100) :: Int
            in (pack . show) percentage <> "%"

getCityStatistics :: Text -> StatDB -> IO (Either Text StatResult)
getCityStatistics city db = do
    isInvalid <- isKeyInvalid db city
    
    if isInvalid
        then pure $ Left "Insufficient or outdated data to perform statistical analysis"
        else do
            -- Extract records from the database
            let stats = map snd $ filter (\(key, _) -> city `isSuffixOf` key) (Map.toList db)
                temps = mapMaybe (readMaybe . unpack . temperature) stats :: [Double]
                anomalies = detectAnomalies stats

            -- Compute statistics
            let statRes = StatResult
                    { Types.mean = roundValue 4 $ Statistics.mean temps
                    , Types.min = minimum temps
                    , Types.max = maximum temps
                    , count = length temps
                    , Types.stdDev = roundValue 4 $ Statistics.stdDev temps
                    , Types.median = Statistics.median temps
                    , Types.mode = Statistics.mode temps
                    , anomaly = parseAnomalies anomalies
                    }
            pure $ Right statRes
    where
        parseAnomalies :: [(Day, Double)] -> Maybe [WeatherAnomaly]
        parseAnomalies anomalies =
            if null anomalies
                then Nothing
                else Just (map (uncurry WeatherAnomaly) anomalies)