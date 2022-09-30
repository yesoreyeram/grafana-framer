package csvFramer

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	u "github.com/yesoreyeram/grafana-framer/framerUtils"
	"github.com/yesoreyeram/grafana-framer/gframer"
)

func TestCsvStringToFrame(t *testing.T) {
	tests := []struct {
		name      string
		csvString string
		options   CSVFramerOptions
		wantFrame *data.Frame
		wantError error
	}{
		{
			name:      "empty csv should return error",
			wantError: errors.New("empty/invalid csv"),
		},
		{
			name:      "valid csv should not return error",
			csvString: strings.Join([]string{`a,b,c`, `1,2,3`, `11,12,13`, `21,22,23`}, "\n"),
			wantFrame: &data.Frame{
				Name: "",
				Fields: []*data.Field{
					data.NewField("a", nil, []*string{u.P("1"), u.P("11"), u.P("21")}),
					data.NewField("b", nil, []*string{u.P("2"), u.P("12"), u.P("22")}),
					data.NewField("c", nil, []*string{u.P("3"), u.P("13"), u.P("23")}),
				},
				RefID: "",
				Meta:  (*data.FrameMeta)(nil),
			},
		},
		{
			name:      "valid csv without headers should not return error",
			csvString: strings.Join([]string{`1,2,3`, `11,12,13`, `21,22,23`}, "\n"),
			options:   CSVFramerOptions{NoHeaders: true},
			wantFrame: &data.Frame{
				Name: "",
				Fields: []*data.Field{
					data.NewField("1", nil, []*string{u.P("1"), u.P("11"), u.P("21")}),
					data.NewField("2", nil, []*string{u.P("2"), u.P("12"), u.P("22")}),
					data.NewField("3", nil, []*string{u.P("3"), u.P("13"), u.P("23")}),
				},
				RefID: "",
				Meta:  (*data.FrameMeta)(nil),
			},
		},
		{
			name:      "framer options should be respected",
			csvString: strings.Join([]string{`a	b	c`, `1	2	3`, `11	12	13`, `21	22	23`}, "\n"),
			options: CSVFramerOptions{FrameName: "foo", Delimiter: "\t", RelaxColumnCount: true, Columns: []gframer.ColumnSelector{
				{Selector: "a", Alias: "A", Type: "number"},
				{Selector: "b", Alias: "b", Type: "string"},
				{Selector: "c", Type: "timestamp_epoch"},
			}},
			wantFrame: &data.Frame{
				Name: "foo",
				Fields: []*data.Field{
					data.NewField("A", nil, []*float64{u.P(float64(1)), u.P(float64(11)), u.P(float64(21))}),
					data.NewField("b", nil, []*string{u.P("2"), u.P("12"), u.P("22")}),
					data.NewField("c", nil, []*time.Time{u.P(time.UnixMilli(3)), u.P(time.UnixMilli(13)), u.P(time.UnixMilli(23))}),
				},
				RefID: "",
				Meta:  (*data.FrameMeta)(nil),
			},
		},
		{
			name:      "relax column count",
			csvString: strings.Join([]string{`a	b	c`, `1	2	3`, `11	12`, `21	22	23`}, "\n"),
			options: CSVFramerOptions{FrameName: "foo", Delimiter: "\t", SkipLinesWithError: true, Columns: []gframer.ColumnSelector{
				{Selector: "a", Alias: "A", Type: "number"},
				{Selector: "b", Alias: "b", Type: "string"},
				{Selector: "c", Type: "timestamp_epoch"},
			}},
			wantFrame: &data.Frame{
				Name: "foo",
				Fields: []*data.Field{
					data.NewField("A", nil, []*float64{u.P(float64(1)), u.P(float64(21))}),
					data.NewField("b", nil, []*string{u.P("2"), u.P("22")}),
					data.NewField("c", nil, []*time.Time{u.P(time.UnixMilli(3)), u.P(time.UnixMilli(23))}),
				},
				RefID: "",
				Meta:  (*data.FrameMeta)(nil),
			},
		},
		{
			name:      "Skip empty lines",
			csvString: strings.Join([]string{`a	b	c`, `1	2	3`, ``, `21	22	23`}, "\n"),
			options: CSVFramerOptions{FrameName: "foo", Delimiter: "\t", Columns: []gframer.ColumnSelector{
				{Selector: "a", Alias: "A", Type: "number"},
				{Selector: "b", Alias: "b", Type: "string"},
				{Selector: "c", Type: "timestamp_epoch_s"},
			}},
			wantFrame: &data.Frame{
				Name: "foo",
				Fields: []*data.Field{
					data.NewField("A", nil, []*float64{u.P(float64(1)), u.P(float64(21))}),
					data.NewField("b", nil, []*string{u.P("2"), u.P("22")}),
					data.NewField("c", nil, []*time.Time{u.P(time.Unix(3, 0)), u.P(time.Unix(23, 0))}),
				},
				RefID: "",
				Meta:  (*data.FrameMeta)(nil),
			},
		},
		{
			name:      "relax column count",
			csvString: strings.Join([]string{`a;b;c`, `1;2;3`, `11;13`, `21;22;23`}, "\n"),
			options: CSVFramerOptions{FrameName: "foo", Delimiter: ";", RelaxColumnCount: true, Columns: []gframer.ColumnSelector{
				{Selector: "a", Alias: "A", Type: "number"},
				{Selector: "b", Alias: "b", Type: "string"},
				{Selector: "c", Type: "string"},
			}},
			wantFrame: &data.Frame{
				Name: "foo",
				Fields: []*data.Field{
					data.NewField("A", nil, []*float64{u.P(float64(1)), u.P(float64(11)), u.P(float64(21))}),
					data.NewField("b", nil, []*string{u.P("2"), u.P("13"), u.P("22")}),
					data.NewField("c", nil, []*string{u.P("3"), nil, u.P("23")}),
				},
				RefID: "",
				Meta:  (*data.FrameMeta)(nil),
			},
		},
		{
			name:      "comment",
			csvString: strings.Join([]string{`# foo`, `a,b,c`, `#01,02,03`, `1,2,3`, `11,12,13`, `21,22,23`, `#`}, "\n"),
			options:   CSVFramerOptions{Comment: "#"},
			wantFrame: &data.Frame{
				Name: "",
				Fields: []*data.Field{
					data.NewField("a", nil, []*string{u.P("1"), u.P("11"), u.P("21")}),
					data.NewField("b", nil, []*string{u.P("2"), u.P("12"), u.P("22")}),
					data.NewField("c", nil, []*string{u.P("3"), u.P("13"), u.P("23")}),
				},
				RefID: "",
				Meta:  (*data.FrameMeta)(nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrame, err := CsvStringToFrame(tt.csvString, tt.options)
			if tt.wantError != nil {
				require.NotNil(t, err)
				assert.Equal(t, tt.wantError, err)
				return
			}
			require.Nil(t, err)
			require.NotNil(t, gotFrame)
			if tt.wantFrame != nil {
				assert.Equal(t, tt.wantFrame, gotFrame)
			}
		})
	}
}
