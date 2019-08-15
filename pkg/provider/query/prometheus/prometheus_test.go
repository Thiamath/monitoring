/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Prometheus query provider tests

package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nalej/derrors"
	"github.com/nalej/monitoring/pkg/provider/query"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)


// Stub implementation of prometheus/client_golang/api/prometheus/v1.API
type fakeAPI struct {
	queries map[string]map[v1.Range][]byte
}

// Query and QueryRange are the only functions we use for now
// Query performs a query for the given time.
func (a *fakeAPI) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	r := v1.Range{
		Start: ts,
	}

	return a.QueryRange(ctx, query, r)
}

// QueryRange performs a query for the given range.
func (a *fakeAPI) QueryRange(ctx context.Context, query string, r v1.Range) (model.Value, error) {
	ranges, found := a.queries[query]
	if !found {
		return nil, derrors.NewNotFoundError("fake provider received unexpected query").WithParams(query)
	}

	res, found := ranges[r]
	if !found {
		return nil, derrors.NewNotFoundError("fake provider received unexpected range").WithParams(query, r)
	}

	var qres queryResult
	err := json.Unmarshal(res, &qres)

	return model.Value(qres.v), err
}

// Empty functions for interface requirements we don't use
var unimplemented = derrors.NewUnimplementedError("fake api stub function not implemented")

// AlertManagers returns an overview of the current state of the Prometheus alert manager discovery.
func (*fakeAPI) AlertManagers(ctx context.Context) (v1.AlertManagersResult, error) {return v1.AlertManagersResult{}, unimplemented}
// CleanTombstones removes the deleted data from disk and cleans up the existing tombstones.
func (*fakeAPI) CleanTombstones(ctx context.Context) error {return unimplemented}
// Config returns the current Prometheus configuration.
func (*fakeAPI) Config(ctx context.Context) (v1.ConfigResult, error) {return v1.ConfigResult{}, unimplemented}
// DeleteSeries deletes data for a selection of series in a time range.
func (*fakeAPI) DeleteSeries(ctx context.Context, matches []string, startTime time.Time, endTime time.Time) error {return unimplemented}
// Flags returns the flag values that Prometheus was launched with.
func (*fakeAPI) Flags(ctx context.Context) (v1.FlagsResult, error) {return v1.FlagsResult{}, unimplemented}
// LabelValues performs a query for the values of the given label.
func (*fakeAPI) LabelValues(ctx context.Context, label string) (model.LabelValues, error) {return nil, unimplemented}
// Series finds series by label matchers.
func (*fakeAPI) Series(ctx context.Context, matches []string, startTime time.Time, endTime time.Time) ([]model.LabelSet, error) {return nil, unimplemented}
// Snapshot creates a snapshot of all current data into snapshots/<datetime>-<rand>
// under the TSDB's data directory and returns the directory as response.
func (*fakeAPI) Snapshot(ctx context.Context, skipHead bool) (v1.SnapshotResult, error) {return v1.SnapshotResult{}, unimplemented}
// Targets returns an overview of the current state of the Prometheus target discovery.
func (*fakeAPI) Targets(ctx context.Context) (v1.TargetsResult, error) {return v1.TargetsResult{}, unimplemented}

// Directly from prometheus/client_golang/api/prometheus/v1/api.go
// queryResult contains result data for a query.
type queryResult struct {
	Type   model.ValueType `json:"resultType"`
	Result interface{}     `json:"result"`

	// The decoded value.
	v model.Value
}

func (qr *queryResult) UnmarshalJSON(b []byte) error {
	v := struct {
		Type   model.ValueType `json:"resultType"`
		Result json.RawMessage `json:"result"`
	}{}

	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	switch v.Type {
	case model.ValScalar:
		var sv model.Scalar
		err = json.Unmarshal(v.Result, &sv)
		qr.v = &sv

	case model.ValVector:
		var vv model.Vector
		err = json.Unmarshal(v.Result, &vv)
		qr.v = vv

	case model.ValMatrix:
		var mv model.Matrix
		err = json.Unmarshal(v.Result, &mv)
		qr.v = mv

	default:
		err = fmt.Errorf("unexpected value type %q", v.Type)
	}
	return err
}

const (
	queryTest1 = `rate(node_cpu_seconds_total{mode="idle"}[120s])`
)

var (
	queryTime1 = time.Unix(1554037344, 922000000).UTC()
	queryTime2 = time.Unix(1553901000, 0).UTC()
	queryTime3 = time.Unix(1553904000, 0).UTC()
	queryStep = time.Duration(1200) * time.Second
)

var queryResults = map[string]map[v1.Range][]byte{
	queryTest1: map[v1.Range][]byte{
		v1.Range{}: []byte(`
{"resultType": "vector","result": [{"metric": {"cpu": "cpu0","instance": "10.240.0.4:9100","mode": "idle"},"value": [1554037350,"0.9125"]},{"metric": {"cpu": "cpu0","instance": "10.240.0.5:9100","mode": "idle"},"value": [1554037350,"0.876333333333605"]},{"metric": {"cpu": "cpu1","instance": "10.240.0.4:9100","mode": "idle"},"value": [1554037350,"0.9133333333331393"]},{"metric": {"cpu": "cpu1","instance": "10.240.0.5:9100","mode": "idle"},"value": [1554037350,"0.8819999999997283"]}]}
		`),
		v1.Range{
			Start: queryTime1,
		}: []byte(`
{"resultType": "vector","result": [{"metric": {"cpu": "cpu0","instance": "10.240.0.4:9100","mode": "idle"},"value": [1554037344.922,"0.9125"]},{"metric": {"cpu": "cpu0","instance": "10.240.0.5:9100","mode": "idle"},"value": [1554037344.922,"0.876333333333605"]},{"metric": {"cpu": "cpu1","instance": "10.240.0.4:9100","mode": "idle"},"value": [1554037344.922,"0.9133333333331393"]},{"metric": {"cpu": "cpu1","instance": "10.240.0.5:9100","mode": "idle"},"value": [1554037344.922,"0.8819999999997283"]}]}
		`),
		v1.Range{
			Start: queryTime2,
			End: queryTime3,
			Step: queryStep,
		}: []byte(`
{"resultType": "matrix", "result": [{"metric": {"cpu": "cpu0", "instance": "10.240.0.4:9100", "mode": "idle"}, "values": [[1553901000, "0.9064999999997477"], [1553902200, "0.9094999999996314"], [1553903400, "0.908833333333314"]]}, {"metric": {"cpu": "cpu0", "instance": "10.240.0.5:9100", "mode": "idle"}, "values": [[1553901000, "0.8698333333333721"], [1553902200, "0.8754999999999806"], [1553903400, "0.8701666666665309"]]}, {"metric": {"cpu": "cpu1", "instance": "10.240.0.4:9100", "mode": "idle"}, "values": [[1553901000, "0.9223333333332752"], [1553902200, "0.9155000000003687"], [1553903400, "0.9121666666668413"]]}, {"metric": {"cpu": "cpu1", "instance": "10.240.0.5:9100", "mode": "idle"}, "values": [[1553901000, "0.8965000000001359"], [1553902200, "0.8935000000002522"], [1553903400, "0.8916666666666667"]]}]}
		`),
	},
	"scalar(sum(node_memory_MemAvailable_bytes))": map[v1.Range][]byte{
		v1.Range{}: []byte(`{"resultType":"scalar","result":[1554037344.922,"18893152256"]}`),
	},
	"scalar(sum(avg_over_time(node_filesystem_free_bytes[600s])))": map[v1.Range][]byte{
		v1.Range{}: []byte(`{"resultType":"scalar","result":[1554037344.922,"294341394022.4"]}`),
	},
	"scalar(irate(services_created_total[2m]) * 60)": map[v1.Range][]byte{
		v1.Range{}: []byte(`{"resultType":"scalar","result":[1554037344.922,"8"]}`),
	},
	"empty": map[v1.Range][]byte{
		v1.Range{}: []byte(`{"resultType":"vector","result":[]}`),
	},
	"nan": map[v1.Range][]byte{
		v1.Range{}: []byte(`{"resultType":"scalar","result":[1234, "NaN"]}`),
	},
}

func timeParse(in string) time.Time {
	res, err := time.Parse(time.RFC3339, in)
	if err != nil {
		return time.Time{}
	}
	return res
}

var _ = ginkgo.Describe("prometheus", func() {

	var provider *PrometheusProvider

	ginkgo.BeforeSuite(func() {
		var derr derrors.Error
		provider, derr = NewProvider(&PrometheusConfig{})
		gomega.Expect(derr).To(gomega.Succeed())

		fakeapi := &fakeAPI{
			queries: queryResults,
		}
		provider.api = fakeapi
	})

	ginkgo.Context("Query", func() {
		ginkgo.It("should execute timed query", func() {
			q := &query.Query{
				QueryString: queryTest1,
				Range: query.QueryRange{
					Start: queryTime1,
				},
			}

			res := &PrometheusResult{
				Type: PrometheusResultVector,
				Values: []*PrometheusResultValue{
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu0",
							"instance": "10.240.0.4:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-31T13:02:24.922Z"),
								Value: "0.9125",
							},
						},
					},
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu0",
							"instance": "10.240.0.5:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-31T13:02:24.922Z"),
								Value: "0.876333333333605",
							},
						},
					},
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu1",
							"instance": "10.240.0.4:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-31T13:02:24.922Z"),
								Value: "0.9133333333331393",
							},
						},
					},
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu1",
							"instance": "10.240.0.5:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-31T13:02:24.922Z"),
								Value: "0.8819999999997283",
							},
						},
					},
				},
			}

			gomega.Expect(provider.Query(context.Background(), q)).To(gomega.Equal(res))
		})

		ginkgo.It("should execute range query", func() {
			q := &query.Query{
				QueryString: queryTest1,
				Range: query.QueryRange{
					Start: queryTime2,
					End: queryTime3,
					Step: queryStep,
				},
			}

			res := &PrometheusResult{
				Type: PrometheusResultMatrix,
				Values: []*PrometheusResultValue{
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu0",
							"instance": "10.240.0.4:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:10:00Z"),
								Value: "0.9064999999997477",
							},
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:30:00Z"),
								Value: "0.9094999999996314",
							},
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:50:00Z"),
								Value: "0.908833333333314",
							},
						},
					},
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu0",
							"instance": "10.240.0.5:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:10:00Z"),
								Value: "0.8698333333333721",
							},
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:30:00Z"),
								Value: "0.8754999999999806",
							},
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:50:00Z"),
								Value: "0.8701666666665309",
							},
						},
					},
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu1",
							"instance": "10.240.0.4:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:10:00Z"),
								Value: "0.9223333333332752",
							},
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:30:00Z"),
								Value: "0.9155000000003687",
							},
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:50:00Z"),
								Value: "0.9121666666668413",
							},
						},
					},
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu1",
							"instance": "10.240.0.5:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:10:00Z"),
								Value: "0.8965000000001359",
							},
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:30:00Z"),
								Value: "0.8935000000002522",
							},
							&PrometheusValue{
								Timestamp: timeParse("2019-03-29T23:50:00Z"),
								Value: "0.8916666666666667",
							},
						},
					},
				},
			}

			gomega.Expect(provider.Query(context.Background(), q)).To(gomega.Equal(res))
		})

		ginkgo.It("should execute non-range non-timed query", func() {
			q := &query.Query{
				QueryString: queryTest1,
			}

			res := &PrometheusResult{
				Type: PrometheusResultVector,
				Values: []*PrometheusResultValue{
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu0",
							"instance": "10.240.0.4:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-31T13:02:30Z"),
								Value: "0.9125",
							},
						},
					},
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu0",
							"instance": "10.240.0.5:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-31T13:02:30Z"),
								Value: "0.876333333333605",
							},
						},
					},
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu1",
							"instance": "10.240.0.4:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-31T13:02:30Z"),
								Value: "0.9133333333331393",
							},
						},
					},
					&PrometheusResultValue{
						Labels: map[string]string{
							"cpu": "cpu1",
							"instance": "10.240.0.5:9100",
							"mode": "idle",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: timeParse("2019-03-31T13:02:30Z"),
								Value: "0.8819999999997283",
							},
						},
					},
				},
			}

			gomega.Expect(provider.Query(context.Background(), q)).To(gomega.Equal(res))

		})

		ginkgo.It("should handle bad query", func() {
			q := &query.Query{
				QueryString: "bad query",
			}
			res, err := provider.Query(context.Background(), q)
			gomega.Expect(res).To(gomega.BeNil())
			gomega.Expect(err).To(gomega.HaveOccurred())
		})

		ginkgo.It("should handle empty query", func() {
			q := &query.Query{
				QueryString: "empty",
			}
			res := &PrometheusResult{
				Type: PrometheusResultVector,
				Values: []*PrometheusResultValue{},
			}
			gomega.Expect(provider.Query(context.Background(), q)).To(gomega.Equal(res))
		})

		ginkgo.It("should handle nan scalar", func() {
			q := &query.Query{
				QueryString: "nan",
			}
			res, err := provider.Query(context.Background(), q)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(res.(*PrometheusResult).GetScalarInt()).To(gomega.Equal(int64(0)))
		})
	})

	// We're not testing the queries itself here, we need integration
	// testing for that. We just test correct template selection
	// and variable substitution in a non-exhaustive way.
	ginkgo.Context("ExecuteTemplate", func() {
		ginkgo.It("should execute template without average", func() {
			gomega.Expect(
				provider.ExecuteTemplate(context.Background(),
					query.TemplateName_Memory + query.TemplateName_Available,
					nil),
				).To(gomega.Equal(int64(18893152256)))
		})

		ginkgo.It("should execute template with small average", func() {
			gomega.Expect(
				provider.ExecuteTemplate(context.Background(),
					query.TemplateName_Memory + query.TemplateName_Available,
					&query.TemplateVars{AvgSeconds: 10}),
				).To(gomega.Equal(int64(18893152256)))
		})

		ginkgo.It("should execute template with larger average", func() {
			gomega.Expect(
				provider.ExecuteTemplate(context.Background(),
					query.TemplateName_Storage + query.TemplateName_Available,
					&query.TemplateVars{AvgSeconds: 600}),
				).To(gomega.Equal(int64(294341394022)))
		})

		ginkgo.It("should execute counter template", func() {
			tname, err := query.GetPlatformTemplateName(query.MetricCreated)
			gomega.Expect(err).To(gomega.Succeed())

			gomega.Expect(
				provider.ExecuteTemplate(context.Background(),
					tname,
					&query.TemplateVars{
						MetricName: "services",
						StatName: "created",
					}),
				).To(gomega.Equal(int64(8)))
		})
	})
})
