package bloomfilter

import (
	"math"
	"testing"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func Test_Plot(t *testing.T) {
	// 创建绘图对象
	p := plot.New()
	// 设置标题和坐标轴标签
	p.Title.Text = "Plot of p = 2⁻ᵏ"
	p.X.Label.Text = "k"
	p.Y.Label.Text = "p"

	points := make(plotter.XYs, 0)
	for k := 0.0; k <= 16; k += 1 {
		points = append(points, plotter.XY{X: k, Y: math.Pow(2, -k)})
	}

	// 创建折线图
	line, err := plotter.NewLine(points)
	if err != nil {
		panic(err)
	}
	p.Add(line)

	// 保存为PNG文件
	if err := p.Save(12*vg.Inch, 8*vg.Inch, "2_power_negative_k.png"); err != nil {
		panic(err)
	}
}
