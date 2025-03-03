package morph

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type deepUnknownTestSampleInput struct {
	T tftypes.Type
	V tftypes.Value
}
type deepUnknownTestSample struct {
	In  deepUnknownTestSampleInput
	Out tftypes.Value
}

func TestDeepUnknown(t *testing.T) {
	samples := map[string]deepUnknownTestSample{
		"string-nil": {
			In: deepUnknownTestSampleInput{
				T: tftypes.String,
				V: tftypes.NewValue(tftypes.String, nil),
			},
			Out: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		"object": {
			In: deepUnknownTestSampleInput{
				T: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"kind":       tftypes.String,
					"apiVersion": tftypes.String,
					"metadata": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":      tftypes.String,
						"namespace": tftypes.String,
						"labels": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
							"app": tftypes.String,
						}},
					}},
				}},
				V: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"kind":       tftypes.String,
					"apiVersion": tftypes.String,
					"metadata": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":      tftypes.String,
						"namespace": tftypes.String,
						"labels": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
							"app": tftypes.String,
						}},
					}},
				}}, map[string]tftypes.Value{
					"kind":       tftypes.NewValue(tftypes.String, "ConfigMap"),
					"apiVersion": tftypes.NewValue(tftypes.String, "v1"),
					"metadata": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":      tftypes.String,
						"namespace": tftypes.String,
						"labels": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
							"app": tftypes.String,
						}},
					}}, map[string]tftypes.Value{
						"name": tftypes.NewValue(tftypes.String, "foo"),
					}),
				}),
			},
			Out: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"kind":       tftypes.String,
				"apiVersion": tftypes.String,
				"metadata": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":      tftypes.String,
					"namespace": tftypes.String,
					"labels": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"app": tftypes.String,
					}},
				}},
			}}, map[string]tftypes.Value{
				"kind":       tftypes.NewValue(tftypes.String, "ConfigMap"),
				"apiVersion": tftypes.NewValue(tftypes.String, "v1"),
				"metadata": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":      tftypes.String,
					"namespace": tftypes.String,
					"labels": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"app": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"name":      tftypes.NewValue(tftypes.String, "foo"),
					"namespace": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"labels": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"app": tftypes.String,
					}}, map[string]tftypes.Value{
						"app": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					}),
				}),
			}),
		},
	}
	for n, s := range samples {
		t.Run(n, func(t *testing.T) {
			rv, err := DeepUnknown(s.In.T, s.In.V, tftypes.AttributePath{})
			if err != nil {
				t.Logf("Conversion failed for sample '%s': %s", n, err)
				t.FailNow()
			}
			if !cmp.Equal(rv, s.Out, cmp.Exporter(func(t reflect.Type) bool { return true })) {
				t.Logf("Result doesn't match expectation for sample '%s'", n)
				t.Logf("\t Sample:\t%s", spew.Sdump(s.In))
				t.Logf("\t Expected:\t%s", spew.Sdump(s.Out))
				t.Logf("\t Received:\t%s", spew.Sdump(rv))
				t.Fail()
			}
		})
	}
}
