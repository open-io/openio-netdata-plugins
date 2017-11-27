package netdata

import(
    "fmt"
    "strings"
)

// Charts -- list of already created charts with the dimensions
var Charts = make(map[string]map[string]bool)

// Buffer -- metric buffer to be sent
var Buffer = make(map[string][]string)

// DimPrefix -- prefix to add to dimensions
var dimPrefix = "openio"

/*
Update - queue a new metric value on a chart
*/
func Update(chart string, dim string, value string) {
	dim = strings.Replace(dim, ".", "_", -1)
	dim = strings.Replace(dim, ":", "_", -1)
	chart = fmt.Sprintf("%s.%s", dimPrefix, strings.Replace(chart, ".", "_", -1))
	chartTitle := strings.ToUpper(strings.Join(strings.Split(chart, "_"), " "))
	if _, e := Charts[chart]; !e {
		createChart(chart, "", chartTitle, "", strings.Split(chart, ".")[1])
		Charts[chart] = make(map[string]bool)
	}
	if _, e := Charts[chart][dim]; !e {
		createChart(chart, "", chartTitle, "", strings.Split(chart, ".")[1])
		createDim(dim)
		Charts[chart][dim] = true
	}

	Buffer[chart]=append(Buffer[chart], fmt.Sprintf("SET %s %s", dim, value))
}

func createChart(chart string, desc string, title string, units string, family string) {
	fmt.Printf("CHART %s '%s' '%s' '%s' '%s'\n", chart, desc, title, units, family)
}

func createDim(dim string) {
	fmt.Printf("DIMENSION %s '%s' absolute\n", dim, dim)
}
