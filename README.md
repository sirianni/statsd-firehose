statsd-firehose
---------------

`statsd-firehose` is a simple load-testing tool intended for stressing statsd
and, more importantly, whatever you have behind statsd. In my case, that's
carbon-cache, whisper, and network-attached SSDs.

    Usage of ./statsd-firehose:
      -statsd="127.0.0.1:8125": Statsd URL
      -packetsize=512: UDP packet size for metrics sent to statsd
      -countcount=50000: Number of individual counters to run
      -countinterval=60: Gauge update interval, in seconds
      -gaugecount=50000: Number of individual gauges to run
      -gaugeinterval=60: Gauge update interval, in seconds
      -tags="source:firehose": Comma-separated list of tags to send with each metrics
      -verbose: Verbose print