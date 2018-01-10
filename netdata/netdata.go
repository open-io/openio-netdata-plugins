package netdata

import(
    "fmt"
    "strings"
)

// Charts -- list of already created charts with the dimensions
var charts = make(map[string]map[string]bool)

// Prefix -- prefix to use for metrics
var Prefix = "openio"

/*
Metric - metric to be sent to buffer
*/
type Metric struct {
	Chart string
	Dim string
	Value string
}

/*
Update - queue a new metric value on a chart
*/
func Update(chart string, dim string, value string, c chan Metric) {
	chart = fmt.Sprintf("%s.%s", Prefix, strings.Replace(chart, ".", "_", -1))
	chartTitle := strings.ToUpper(strings.Join(strings.Split(chart, "_"), " "))
	if _, e := charts[chart]; !e {
		createChart(chart, "", chartTitle, "")
		charts[chart] = make(map[string]bool)
	}
	if _, e := charts[chart][dim]; !e {
		createChart(chart, "", chartTitle, "")
		createDim(dim)
		charts[chart][dim] = true
	}

    c <- Metric{
        Chart: chart,
        Dim: dim,
        Value: value,
    }
}

func createChart(chart string, desc string, title string, units string) {
	fmt.Printf("CHART %s '%s' '%s' '%s' '%s'\n", chart, desc, title, units, getFamily(chart))
}

func getFamily(chart string) string {
    families := map[string]string {
        "req": "Request",
        "rep": "Response",
        "score": "Score",
        "byte": "Capacity",
        "inodes": "Inodes",
        "cnx": "Connections",
    }

    chart = strings.Split(chart, ".")[1]
    for k, v := range families {
        if strings.HasPrefix(chart, k) {
            return v
        }
    }
    return "Misc"
}

func createDim(dim string) {
	fmt.Printf("DIMENSION %s '%s' absolute\n", dim, dim)
}
