package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"

	"os"

	"github.com/davvo/mercator"
	"github.com/fogleman/gg"
	"github.com/paulmach/go.geojson"
	yaml "gopkg.in/yaml.v2"
)

const (
	correct        = 10.0
	pointR         = 1
	picWidth       = 1366
	picHeight      = 1024
	scale          = 1.0
	stylesName     = "style.yml"
	defFillColor   = "#757575AA"
	defBorderColor = "#0f0f0fFF"
	defLineWidth   = 0.5
)

var fileName *string

type style struct {
	AdLevel []adminLevel `yaml:"admin_level"`
	Lines   lines        `yaml:"lines"`
}

type adminLevel struct {
	Rank        int     `yaml:"rank"`
	BorderColor string  `yaml:"borderColor"`
	FillColor   string  `yaml:"fillColor"`
	LineWidth   float64 `yaml:"lineWidth"`
}

type lines struct {
	Road   string  `yaml:"road"`
	RWidth float64 `yaml:"roadWidth"`
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	fileName = flag.String("path", "map.geojson", "Choose GeoJSON file")
	flag.Parse()
	fmt.Printf("\nИмя указанного файла: %s\n", *fileName)

	file, _ := os.Open(*fileName)
	defer file.Close()
	stat, _ := file.Stat()
	buf := make([]byte, stat.Size())
	file.Read(buf)

	fc, _ := geojson.UnmarshalFeatureCollection(buf)

	styles := readStyles(stylesName)

	dc := gg.NewContext(picWidth, picHeight)
	dc.SetFillRule(gg.FillRuleEvenOdd)
	dc.InvertY()

	for _, feature := range fc.Features {
		switch feature.Geometry.Type {

		case "Polygon":
			dc.NewSubPath()
			for _, geom := range feature.Geometry.Polygon {
				for _, point := range geom {
					x, y := mercator.LatLonToPixels(point[1], point[0], 1)
					x, y = scalePic(x, y)
					dc.LineTo(x, y)
				}
				dc.ClosePath()
				setStyle(dc, styles, feature.Properties)
			}

		case "LineString":
			dc.NewSubPath()
			for _, geom := range feature.Geometry.LineString {
				x, y := mercator.LatLonToPixels(geom[1], geom[0], 1)
				x, y = scalePic(x, y)
				dc.LineTo(x, y)
				dc.MoveTo(x, y)
			}
			setLineStyle(dc, styles, feature.Properties)

		case "Point":
			x, y := mercator.LatLonToPixels(feature.Geometry.Point[1], feature.Geometry.Point[0], 1)
			x, y = scalePic(x, y)
			dc.DrawPoint(x, y, pointR)
			setStyle(dc, styles, feature.Properties)

		case "MultiPolygon":
			for _, geom := range feature.Geometry.MultiPolygon {
				for _, poly := range geom {
					newSubPoly := true
					for _, point := range poly {
						x, y := mercator.LatLonToPixels(point[1], point[0], 3)
						x, y = scalePic(x, y)
						if newSubPoly {
							dc.MoveTo(x, y)
							newSubPoly = false
						} else {
							dc.LineTo(x, y)
						}
					}
				}
				dc.ClosePath()
				setStyle(dc, styles, feature.Properties)
			}
		}
	}
	dc.SavePNG("out.png")
}

////////////////////////////////////////////////////////////////////////////////

func readStyles(styleFile string) style {
	st := style{}
	file, _ := ioutil.ReadFile(styleFile)
	_ = yaml.Unmarshal(file, &st)
	return st
}

func scalePic(x, y float64) (float64, float64) {
	x = x*scale + correct
	y = y*scale + correct
	return x, y
}

func setLineStyle(dc *gg.Context, st style, prop map[string]interface{}) {
	err := prop["road"]
	if err != nil {
		roadValue, _ := strconv.ParseBool(prop["road"].(string))
		if roadValue {
			dc.SetHexColor(st.Lines.Road)
			dc.SetLineWidth(st.Lines.RWidth)
		} else {
			dc.SetHexColor(defBorderColor)
			dc.SetLineWidth(defLineWidth)
		}
		dc.Stroke()
	}
}

func setStyle(dc *gg.Context, st style, prop map[string]interface{}) {
	err := prop["admin_level"]
	if err != nil {
		level, _ := strconv.Atoi(prop["admin_level"].(string))
		fColor := st.AdLevel[level].FillColor
		bColor := st.AdLevel[level].BorderColor
		lWidth := st.AdLevel[level].LineWidth
		if fColor != "" {
			dc.SetHexColor(fColor)
		} else {
			dc.SetHexColor(defFillColor)
		}
		dc.FillPreserve()
		if bColor != "" {
			dc.SetHexColor(bColor)
		} else {
			dc.SetHexColor(defBorderColor)
		}
		if lWidth != 0.0 {
			dc.SetLineWidth(lWidth)
		} else {
			dc.SetLineWidth(defLineWidth)
		}
	} else {
		dc.SetHexColor(defFillColor)
		dc.FillPreserve()
		dc.SetHexColor(defBorderColor)
		dc.SetLineWidth(defLineWidth)
	}
	dc.Stroke()
}
