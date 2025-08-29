<div align="center">
<h1>Zephyr 🌲</h1>
    
<h6><i>real-time weather forecast service</i></h6>

[![](https://github.com/ceticamarco/zephyr/actions/workflows/docker.yml/badge.svg)](https://github.com/ceticamarco/zephyr/actions/workflows/docker.yml)
[![](https://github.com/ceticamarco/zephyr/actions/workflows/tests.yml/badge.svg)](https://github.com/ceticamarco/zephyr/actions/workflows/tests.yml)

</div>

Zephyr is a lightweight, real-time HTTP service that provides a simple way to gather
weather data and apply statistical analysis to past meteorological records. It is
written in Go using `net/http` and [OpenWeatherMap](https://openweathermap.org/)
for weather data.

I've built this service out of frustration with existing weather platforms
cluttered with ads, paywalls, clickbait content and other unnecessary distractions.
Zephyr is designed to be simple, fast and efficient, providing only the
weather data of a given location, without any additional nonsense.

This service communicates through a JSON API, making it suitable for
any kind of internet-based project or device. I already use it on a widget
on my phone, on my terminal, on the tmux's status bar, in a couple of
smart bedside clocks I've built and as a standalone web app.

## Weather
As state before, Zephyr talks via HTTP using the JSON format. Therefore, you
can query it using any HTTP client of your choice. Below you can find some examples
using `curl`:

```sh
curl -s 'http://127.0.0.1:3000/weather/milan' | jq
```

which yield the following:

```json
{
  "date": "Friday, 2025/08/29",
  "temperature": "18°C",
  "min": "18°C",
  "max": "24°C",
  "condition": "Clouds",
  "feelsLike": "18°C",
  "emoji": "☁️",
  "alerts": [
    {
      "event": "Yellow Thunderstorm Warning",
      "startDate": "Friday, 2025/08/29 2:00 AM",
      "endDate": "Friday, 2025/08/29 11:59 PM",
      "description": "Moderate intensity weather phenomena expected EASTERN ALPINE AND PRE-ALPINE SECTOR"
    },
    {
      "event": "Yellow Thunderstorm Warning",
      "startDate": "Saturday, 2025/08/30 12:00 AM",
      "endDate": "Saturday, 2025/08/30 5:59 AM",
      "description": "Moderate intensity weather phenomena expected"
    },
    {
      "event": "Orange Thunderstorm Warning",
      "startDate": "Friday, 2025/08/29 9:00 AM",
      "endDate": "Friday, 2025/08/29 11:59 PM",
      "description": "Severe weather expected"
    }
  ]
}
```

To get the results in imperial units, you can append the `i` query parameter to the
URL:

```sh
curl -s 'http://127.0.0.1:3000/weather/milan?i' | jq
```

which yields:

```json
{
  "date": "Friday, 2025/08/29",
  "temperature": "50°F",
  "min": "50°F",
  "max": "56°F",
  "condition": "Clouds",
  "feelsLike": "50°F",
  "emoji": "☁️",
  "alerts": [
    {
      "event": "Yellow Thunderstorm Warning",
      "startDate": "Friday, 2025/08/29 2:00 AM",
      "endDate": "Friday, 2025/08/29 11:59 PM",
      "description": "Moderate intensity weather phenomena expected EASTERN ALPINE AND PRE-ALPINE SECTOR"
    },
    {
      "event": "Yellow Thunderstorm Warning",
      "startDate": "Saturday, 2025/08/30 12:00 AM",
      "endDate": "Saturday, 2025/08/30 5:59 AM",
      "description": "Moderate intensity weather phenomena expected"
    },
    {
      "event": "Orange Thunderstorm Warning",
      "startDate": "Friday, 2025/08/29 9:00 AM",
      "endDate": "Friday, 2025/08/29 11:59 PM",
      "description": "Severe weather expected"
    }
  ]
}
```

## Metrics
The `/metrics/:city` endpoint provides environmental metrics for a given city:

```sh
curl -s 'http://127.0.0.1:3000/metrics/taipei' | jq
```

which yields:

```json
{
  "humidity": "23%",
  "pressure": "1015 hPa",
  "dewPoint": "6°C",
  "uvIndex": "4",
  "visibility": "10km"
}
```

As in the previous example, you can append the `i` query parameter to get results
in imperial units.

## Wind
The `/wind/:city` endpoint provides wind related information(such as speed and direction) for a given city:

```sh
curl -s 'http://127.0.0.1:3000/wind/bolzano' | jq
```

which yields:

```json
{
  "arrow": "⬆️",
  "direction": "S",
  "speed": "13.0 km/h"
}
```
As in the previous examples, you can append the `i` query parameter to get results
in imperial units.

## Forecast

### Daily
The `/forecast/:city` endpoint allows you to get the weather forecast of the
next 4 days. For example:

```sh
curl -s 'http://127.0.0.1:3000/forecast/Yakutsk' | jq
```

which yields:

```json
{
  "forecast": [
    {
      "date": "Tuesday, 2025/05/06",
      "min": "-2°C",
      "max": "6°C",
      "condition": "Rain",
      "emoji": "🌧️",
      "feelsLike": "0°C",
      "wind": {
        "arrow": "↗️",
        "direction": "SSW",
        "speed": "14.7 km/h"
      },
      "rainProbability": "100%"
    },
    {
      "date": "Wednesday, 2025/05/07",
      "min": "2°C",
      "max": "9°C",
      "condition": "Snow",
      "emoji": "☃️",
      "feelsLike": "7°C",
      "wind": {
        "arrow": "↘️",
        "direction": "NNW",
        "speed": "13.9 km/h"
      },
      "rainProbability": "100%"
    }
  ]
}
```

As in the previous examples, you can append the `i` query parameter to get results
in imperial units.

### Hourly
You can also get the hourly forecast of a time window of 9 hours by appending
the `h`(hourly) query parameter to the URL:

```sh
curl -s 'http://127.0.0.1:3000/forecast/tapei?h' | jq
```

```json
{
  "forecast": [
    {
      "time": "2:00 PM",
      "temperature": "26°C",
      "condition": "Clouds",
      "emoji": "☁️",
      "wind": {
        "arrow": "↘️",
        "direction": "NW",
        "speed": "23.3 km/h"
      },
      "rainProbability": "0%"
    },
    {
      "time": "3:00 PM",
      "temperature": "27°C",
      "condition": "Clouds",
      "emoji": "☁️",
      "wind": {
        "arrow": "↘️",
        "direction": "NW",
        "speed": "20.2 km/h"
      },
      "rainProbability": "0%"
    }
  ]
}
```

As in the previous examples, you can append the `i` query parameter to get results
in imperial units(**tip**: you can mix both parameter using `&`).

## Moon

The `/moon` endpoint provides the current moon phase and its emoji representation:

```sh
curl -s 'http://127.0.0.1:3000/moon' | jq 
```

will yield

```json
{
  "icon": "🌘",
  "phase": "Waning Crescent",
  "percentage": "44%"
}
```

> [!NOTE]
> To convert OpenWeatherMap's moon phase value to the illumination percentage, 
> I've used the following formula:
> 
> $$
>   \sin(\pi \theta)^2 \times 100
> $$

## Statistical analysis
In addition to the weather data, Zephyr also provides statistical analysis of past
meteorological records. This is done through the `/stats/:city` endpoint, which
returns additional information about the weather of the previous days such as
the average temperature, the maximum and minimum temperatures, the standard deviation,
the median and the mode.

This endpoint becomes available only after the service has collected enough 
**updated** data for a given city. In particular, the services will require
**at least** two weather records **within the last 48 hours**. If these two
conditions aren't met, the service will refuse to provide statistical data.

After enough data has been collected in the in-memory database, you will be
able to query the statistics endpoint like this:

```sh
$ curl -s 'http://127.0.0.1:3000/stats/berlin' | jq
```

which yields:

```json
{
  "min": "25°C",
  "max": "25°C",
  "count": 30,
  "mean": "25°C",
  "stdDev": "0.1821°C",
  "median": "25°C",
  "mode": "25°C",
  "anomaly": null
}
```
The service is also able to detect anomalies in the temperature data using a built-in statistical model. 
For instance, two temperature spikes, such as `+34°C` and `-15°C`, with a mean of `25°C` and a standard deviation of `0.2°C`,
will be flagged as outliers by the model and will be reported as such:

```json
{
  "min": "-15°C",
  "max": "34°C",
  "count": 32,
  "mean": "24°C",
  "stdDev": "7.1864°C",
  "median": "25°C",
  "mode": "25°C",
  "anomaly": [
    {
      "date": "Sunday, 2025/06/01",
      "temperature": "-15°C"
    },
    {
      "date": "Wednesday, 2025/05/28",
      "temperature": "34°C"
    }
  ]
}
```

### Anomaly Detection
The anomaly detection algorithm is based on a modified version of the
[Z-Score](https://en.wikipedia.org/wiki/Standard_score) algorithm, which uses the
[Median Absolute Deviation](https://en.wikipedia.org/wiki/Median_absolute_deviation) to measure the variability
in a given sample of quantitative data. The algorithm can be summarized as follows(let $X$ be the sample):


$$
    \tilde{x} = \text{median}({X})
$$

Compute the median absolute deviation

$$
  \text{MAD} = \text{median}\{ |x_i - \tilde{x}| : \forall i = 0, \dots, n-1 \}
$$

Compute the (modified)Z-score

$$
  z_i = \frac{0.6745 (x_i - \tilde{x})}{\text{MAD}}
  \quad \forall i = 0, \dots, n-1
$$

Flag $x_i$ as an outlier if:

$$
    |z_i| > 4.5
$$

and

$$
    |x_i-\tilde{x}| \geq 8
$$

Here, $\Phi^{-1}(3/4) = \Phi^{-1}(0.75) \approx 0.6745$ reflects the fact
that 75% of values lie within $\approx 0.6745$ standard deviation, 4.5 represent a fixed threshold value and 8 represent the minimum absolute deviation that a value
must have from the median to be considered an outlier.

These constants have been fine-tuned to work well with the weather data of
a wide range of climates and to ignore daily temperature fluctuations while
still being able to detect significant anomalies.

According to the Q-Q plots, daily temperatures collected over a time window of no more than 1/2 months
but no less than a week, *should* follow a normal distribution.

> [!IMPORTANT]
> The anomaly detection algorithm works under the assumption that the weather data
> is normally distributed(at least roughly), this might not be the case on datasets
> with a very small number of samples(e.g. few days of data) or with a large
> number of samples(e.g. multi-seasonal data). 

The algorithm works quite well when these conditions are met, and even with real world data,
the results were quite satisfactory. However, if it
start to produce false positives, you will need to dump the whole in-memory
database and start from scratch. I recommend to do this at every change of season.

## Embedded Cache System
To minimize the amount of requests sent to the OpenWeatherMap API, Zephyr provides a built-in,
in-memory cache data structure that stores fetched weather data. Each time a client requests
weather data for a given location, the service will first check if it's already available on the cache.
If it is found, the cached value will be returned, otherwise a new request will be sent to the OpenWeatherMap API
and the response will be returned to the client and stored in the cache for future use. Each cache entry
is valid for a fixed amount of time, which can be configured by setting the `ZEPHYR_CACHE_TTL` environment variable. Once
a cached entry expires, Zephyr will retrieve a new value from the OpenWeatherMap API and update the cache accordingly.

The cache system significantly improves the performance of the service by decreasing its latency. Additionally, it
also helps to reduce the number of API calls made to the OpenWeatherMap servers, which is quite important
if you are using their free tier.

## Configuration
Zephyr requires the following environment variables to be set:

| Variable             | Meaning                                |
|----------------------|----------------------------------------|
| `ZEPHYR_ADDR`        | Listen address                         |
| `ZEPHYR_PORT`        | Listen port                            |
| `ZEPHYR_TOKEN`       | OpenWeatherMap API key                 |
| `ZEPHYR_CACHE_TTL`   | Cache time-to-live(expressed in hours) |

Each value must be set _before_ launching the application. If you plan to deploy Zephyr using
Docker, you can specify these variables in the `compose.yml` file.

You will also need an OpenWeatherMap API key, you can get one for free by following
the instructions [listed on their website](https://openweathermap.org/api).

> [!NOTE]
> Zephyr is designed to work with OpenWeatherMap's free tier. As long as you
> stay within the daily limits of 1,000 requests, you won't need to pay.

## Deploy
Zephyr can be deployed using Docker by just issuing the following command:

```sh
docker compose up -d
```

This will build the container image and start the service in detached mode. By default,
the service will be available at `http://127.0.0.1:3000`, but you can easily change this property
but editing the `compose.yml` as stated above.

## Unit tests
You can run the unit tests by issuing the following command:

```sh
 go test ./... -v
 ```

## License
This software is released under the GPLv3 license. You can find a copy of the license with this repository or by visiting the [following page](https://choosealicense.com/licenses/gpl-3.0/).
