statsd-firehose
---------------

`statsd-firehose` is a simple load-testing tool intended for stressing statsd
and, more importantly, whatever you have behind statsd. In my case, that's
carbon-cache, whisper, and network-attached SSDs.

    Usage of ./statsd-firehose:
      -countcount int
      Number of individual counters to run (default 10000)
      -countfreq int
      How many times to update each individual count per interval (default 1)
      -countinterval int
      Gauge update interval, in seconds (default 1)
      -distcount int
      Number of individual distributions to run (default 10000)
      -distfreq int
      How many times to update each individual distribution per interval (default 1)
      -distinterval int
      Distribution update interval, in seconds (default 1)
      -gaugecount int
      Number of individual gauges to run (default 10000)
      -gaugefreq int
      How many times to update each individual gauge per interval (default 1)
      -gaugeinterval int
      Gauge update interval, in seconds (default 1)
      -namespace string
      Namespace for firehose metrics (default "firehose")
      -packetsize int
      UDP packet size for metrics sent to statsd (default 512)
      -statsd string
      Statsd URL (default "127.0.0.1:8125")
      -tags string
      Comma-separated list of tags to send with each metrics (default "source:firehose")
      -verbose
      Verbose print