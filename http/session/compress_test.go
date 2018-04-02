package session

import (
	"bytes"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestCompress(t *testing.T) {
	spec.Run(t, "Compress", testCompress, spec.Report(report.Terminal{}))
}

func testCompress(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("can round-trip a byte array", func() {
		b := bytes.NewBuffer(nil)
		GzipWrite(b, []byte("hello"))
		Expect(b.String()).NotTo(Equal("hello"))
		b2 := bytes.NewBuffer(nil)
		GunzipWrite(b2, b.Bytes())
		Expect(b2.String()).To(Equal("hello"))
	})
}
