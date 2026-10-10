package video

type conversionSlots chan struct{}

func newConversionSlots(capacity int) conversionSlots {
	return make(conversionSlots, capacity)
}

func (slots conversionSlots) tryAcquire() bool {
	select {
	case slots <- struct{}{}:
		return true
	default:
		return false
	}
}

func (slots conversionSlots) release() {
	<-slots
}
