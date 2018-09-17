package netdata

import (
	"fmt"
	"strings"
	"time"
)

type Algorithm string

const (
	AbsoluteAlgorithm    Algorithm = "absolute"
	IncrementalAlgorithm Algorithm = "incremental"
)

type Chart struct {
	ID       string
	Type     string
	Name     string
	Title    string
	Units    string
	Family   string
	Category string

	dimensions      map[string]Dimension
	dimensionsIndex []string

	refresh bool
}

func NewChart(chartType, id, name, title, units, family, category string) *Chart {
	return &Chart{
		Type:       chartType,
		ID:         id,
		Name:       name,
		Title:      title,
		Units:      units,
		Family:     family,
		Category:   category,
		dimensions: make(map[string]Dimension),
		refresh:    true,
	}
}

type Dimension struct {
	id string

	name string

	algorithm Algorithm
}

func (d *Dimension) create() string {
	return fmt.Sprintf("DIMENSION '%v' '%v' %v", d.id, d.name, d.algorithm)
}

func (d *Dimension) set(value string) string {
	return fmt.Sprintf("SET '%s' = %s", d.id, value)
}

func (c *Chart) AddDimension(id, name string, algorithm Algorithm) {
	c.dimensionsIndex = append(c.dimensionsIndex, id)

	c.dimensions[id] = Dimension{
		id:        id,
		name:      name,
		algorithm: algorithm,
	}
}

func (c *Chart) create(out Writer) {
	chartCreate := fmt.Sprintf("CHART %s.%s '%s' '%s' '%s' '%s' '%s'", c.Type, c.ID, c.Name, c.Title, c.Units, c.Family, c.Category)
	dimensionsCreate := []string{}
	for _, dimID := range c.dimensionsIndex {
		dim := c.dimensions[dimID]
		dimensionsCreate = append(dimensionsCreate, dim.create())
	}
	out.Printf("%v\n%v\n", chartCreate, strings.Join(dimensionsCreate, "\n"))
	c.refresh = false
}

func (c *Chart) Update(data map[string]string, interval time.Duration, out Writer) bool {
	var updatedDimensions []string
	for _, dimID := range c.dimensionsIndex {
		if c.refresh {
			c.create(out)
		}
		if value, ok := data[dimID]; ok {
			dim := c.dimensions[dimID]
			updatedDimensions = append(updatedDimensions, dim.set(value))
		}
	}

	if len(updatedDimensions) != 0 {
		out.Printf("BEGIN %s.%s\n%s\nEND\n", c.Type, c.ID, strings.Join(updatedDimensions, "\n"))
		return true
	}

	return false
}

type Charts map[string]*Chart
