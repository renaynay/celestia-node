package metrics

type Histogram struct {

}

func NewHistogram() *Histogram {
	durRecorder, _ := meter.SyncInt64().Histogram(
		"some.prefix.histogram",
		instrument.WithUnit("microseconds"),
		instrument.WithDescription("TODO"),
	)
}

func (h *Histogram) Record() {

}
