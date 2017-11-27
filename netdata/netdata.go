package netdata

import(
    "fmt"
    "strings"
)

// Charts -- list of already created charts with the dimensions
var Charts = make(map[string]map[string]bool)
var prefix = "openio"

/*
Update - queue a new metric value on a chart
*/
func Update(chart string, dim string, value string) {
	dim = strings.Replace(dim, ".", "_", -1)
	dim = strings.Replace(dim, ":", "_", -1)
	chart = fmt.Sprintf("%s.%s", prefix, strings.Replace(chart, ".", "_", -1))
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
    send(chart, dim, value)
}

func send(chart string, dim string, value string) {
    fmt.Printf("BEGIN %s\n", chart)
    fmt.Printf("SET %s %s\n", dim, value)
    fmt.Println("END")
}

func createChart(chart string, desc string, title string, units string, family string) {
	fmt.Printf("CHART %s '%s' '%s' '%s' '%s'\n", chart, desc, title, units, family)
}

func createDim(dim string) {
	fmt.Printf("DIMENSION %s '%s' absolute\n", dim, dim)
}
