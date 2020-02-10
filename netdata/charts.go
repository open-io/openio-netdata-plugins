// OpenIO netdata collectors
// Copyright (C) 2019 OpenIO SAS
//
// This library is free software; you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3.0 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Lesser General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public
// License along with this program. If not, see <http://www.gnu.org/licenses/>.

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

type Charts map[string]*Chart

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
	var updated []string
	for dim, value := range data {
		if d, ok := c.dimensions[dim]; ok {
			updated = append(updated, d.set(value))
		}
		if strings.HasPrefix(dim, c.ID+"_") {
			dim = strings.TrimPrefix(dim, c.ID+"_")
			if _, ok := c.dimensions[dim]; !ok {
				c.AddDimension(dim, dim, AbsoluteAlgorithm)
				c.refresh = true
			}
			dm := c.dimensions[dim]
			updated = append(updated, dm.set(value))
		}
	}
	if c.refresh {
		c.create(out)
	}

	if len(updated) > 0 {
		out.Printf("BEGIN %s.%s\n%s\nEND\n", c.Type, c.ID, strings.Join(updated, "\n"))
		return true
	}

	return false
}
