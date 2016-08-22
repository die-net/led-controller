package main

type AudioMv struct {
	Count float32 `json:"count,omitempty"`
	Min   float32 `json:"min"`
	Avg   float32 `json:"avg"`
	Max   float32 `json:"max"`
}

func (mv AudioMv) MovingAverage(n AudioMv, samples float32) AudioMv {
	mult := samples - 1
	mv.Count = (mv.Count*mult + n.Count) / samples
	mv.Avg = (mv.Avg*mult + n.Avg) / samples
	if n.Min < mv.Min {
		mv.Min = n.Min
	} else {
		mv.Min = (mv.Min*mult + n.Min) / samples
	}
	if n.Max > mv.Max {
		mv.Max = n.Max
	} else {
		mv.Max = (mv.Max*mult + n.Max) / samples
	}
	return mv
}

func (mv AudioMv) Amplitude() int {
	return int(mv.Max - mv.Min + 0.5)
}
