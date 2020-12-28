package query

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	singleTenantHeader    = map[string][]string{"X-Scope-OrgID": {"scp"}, "Cache-Control": {"no-store"}}
	multiTenantHeader     = map[string][]string{"X-Scope-OrgID": {"k8s-kube-mj2drd-iad10|k8s-genuine-cat|k8s-gstage1|k8s-play3|k8s-app-play1|k8s-control1|k8s-app-stage1|k8s-app-prod1"}, "Cache-Control": {"no-store"}}
	instantQueryDuration  *prometheus.HistogramVec
	queryRangeDuration    *prometheus.HistogramVec
	fourHourRangeDuration *prometheus.HistogramVec
)

func RegisterMetrics() {
	instantQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "avalanche_query",
		Name:      "request_duration_seconds",
		Help:      "Time (in seconds) for query to complete.",
		Buckets:   []float64{.05, .1, .15, .18, .2, .22, .25, .28, .3, .5, 0.75, 1, 1.5, 2},
	}, []string{"tenant_type", "status_code"})
	queryRangeDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "avalanche_query_range",
		Name:      "request_duration_seconds",
		Help:      "Time (in seconds) for query to complete.",
		Buckets:   []float64{15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30},
	}, []string{"tenant_type", "status_code"})
	fourHourRangeDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "avalanche_four_hour_range",
		Name:      "request_duration_seconds",
		Help:      "Time (in seconds) for query to complete.",
		Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 0.75, 1, 1.5, 2, 3, 4, 5, 10, 25, 50, 100},
	}, []string{"tenant_type", "status_code"})
}

func Run(queryInterval int, cortexFrontend string) {
	runInstantQuery(queryInterval, cortexFrontend)
	runTwoDayQuery(cortexFrontend)
	runFourHourQuery(cortexFrontend)
}

func runInstantQuery(queryInterval int, cortexFrontend string) {
	go func() {
		queryTick := time.NewTicker(time.Duration(queryInterval) * time.Second)
		for tick := range queryTick.C {
			s1, d1 := doPromQuery(fmt.Sprintf("{__name__=~\"avalanche_metric_mmmmm_0_1\"}"), singleTenantHeader, cortexFrontend)
			instantQueryDuration.WithLabelValues("single", strconv.Itoa(s1)).Observe(d1)
			s2, d2 := doPromQuery(fmt.Sprintf("{__name__=~\"avalanche_metric_mmmmm_0_1\"}"), multiTenantHeader, cortexFrontend)
			instantQueryDuration.WithLabelValues("multi", strconv.Itoa(s2)).Observe(d2)
			fmt.Printf("%v: running query\n", tick)
		}
	}()
}

func runTwoDayQuery(cortexFrontend string) {
	qEnd := time.Now().Unix()
	qStart := qEnd - (2 * 24 * 60 * 60)
	go func() {
		for {
			q := "{__name__=~\"avalanche_metric_mmmmm_0_1\"}"
			fmt.Printf("running query %s\n", q)
			s1, d1 := doPromRangeQuery(q, singleTenantHeader, cortexFrontend, qStart, qEnd)
			queryRangeDuration.WithLabelValues("single", strconv.Itoa(s1)).Observe(d1)
			s2, d2 := doPromRangeQuery(q, multiTenantHeader, cortexFrontend, qStart, qEnd)
			queryRangeDuration.WithLabelValues("multi", strconv.Itoa(s2)).Observe(d2)
		}
	}()
}

func runFourHourQuery(cortexFrontend string) {
	qEnd := time.Now().Unix()
	qStart := qEnd - (4 * 60 * 60)
	go func() {
		for {
			q := "{__name__=~\"avalanche_metric_mmmmm_0_1\"}"
			fmt.Printf("running query %s\n", q)
			s1, d1 := doPromRangeQuery(q, singleTenantHeader, cortexFrontend, qStart, qEnd)
			fourHourRangeDuration.WithLabelValues("single", strconv.Itoa(s1)).Observe(d1)
			s2, d2 := doPromRangeQuery(q, multiTenantHeader, cortexFrontend, qStart, qEnd)
			fourHourRangeDuration.WithLabelValues("multi", strconv.Itoa(s2)).Observe(d2)
		}
	}()
}

func doPromQuery(query string, header map[string][]string, cortexFrontend string) (int, float64) {
	request, err := http.NewRequest("GET", fmt.Sprintf("http://%s/prometheus/api/v1/query", cortexFrontend), nil)
	if err != nil {
		// do something
	}
	request.Header = header
	q := request.URL.Query()
	q.Add("query", query)
	request.URL.RawQuery = q.Encode()
	client := &http.Client{}
	start := time.Now()
	resp, err := client.Do(request)
	fmt.Print(err)
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if data != nil {
	}
	return resp.StatusCode, time.Since(start).Seconds()
}

func doPromRangeQuery(query string, header map[string][]string, cortexFrontend string, qStart int64, qEnd int64) (int, float64) {
	request, err := http.NewRequest("GET", fmt.Sprintf("http://%s/prometheus/api/v1/query_range", cortexFrontend), nil)
	if err != nil {
		// do something
	}
	request.Header = header
	q := request.URL.Query()
	q.Add("query", query)
	q.Add("start", strconv.Itoa(int(qStart)))
	q.Add("end", strconv.Itoa(int(qEnd)))
	q.Add("step", "1m")
	request.URL.RawQuery = q.Encode()
	client := &http.Client{}
	start := time.Now()
	resp, err := client.Do(request)
	fmt.Print(err)
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if data != nil {
	}
	return resp.StatusCode, time.Since(start).Seconds()
}
