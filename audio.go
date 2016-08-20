package main

type AudioMv struct {
	Count float32 `json:"count,omitempty"`
	Min   float32 `json:"min"`
	Avg   float32 `json:"avg"`
	Max   float32 `json:"max"`
}

func (mv AudioMv) MovingAverage(n AudioMv) AudioMv {
	mv.Count = (mv.Count*255 + n.Count) / 256
	mv.Avg = (mv.Avg*255 + n.Avg) / 256
	if n.Min < mv.Min {
		mv.Min = n.Min
	} else {
		mv.Min = (mv.Min*255 + n.Min) / 256
	}
	if n.Max > mv.Max {
		mv.Max = n.Max
	} else {
		mv.Max = (mv.Max*255 + n.Max) / 256
	}
	return mv
}

func (mv AudioMv) Amplitude() int {
	return int(mv.Max - mv.Min + 0.5)
}
