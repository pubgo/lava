package resty

// https://github.com/go-resty/resty
// https://github.com/imroc/req
// https://github.com/sony/gobreaker
// newCb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
//			Name:        requestURL,
//			MaxRequests: 5,                // MaxRequests pass through cb when state if half-open
//			Interval:    time.Minute,      // Reset counter in open
//			Timeout:     time.Second * 10, // change to half-open when open
//			ReadyToTrip: func(counts gobreaker.Counts) bool {
//				return counts.ConsecutiveFailures >= 10
//			},
//			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
//				b.logger.print(fmt.Sprintf("httpclient.circuitbreaker.OnStateChange from %s to %s", from.String(), to.String()))
//			},
//		})
