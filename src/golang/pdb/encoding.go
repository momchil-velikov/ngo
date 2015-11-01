package pdb

import "io"

type Encoder struct {
	w   io.Writer
	buf []byte
}

func (e *Encoder) Init(w io.Writer) *Encoder {
	e.w = w
	e.buf = nil
	return e
}

func (e *Encoder) WriteNum(n uint64) error {
	e.buf = e.buf[:0]
	if n == 0 {
		e.buf = append(e.buf, 0)
	} else {
		for n != 0 {
			b := byte(n & 0x7f)
			n >>= 7
			if n != 0 {
				b |= 0x80
			}
			e.buf = append(e.buf, b)
		}
	}
	_, err := e.w.Write(e.buf)
	return err
}

func (e *Encoder) WriteString(s string) error {
	n := len(s)
	if n == 0 {
		return e.WriteNum(0)
	}
	err := e.WriteNum(uint64(n))
	if err != nil {
		return err
	}
	e.buf = append(e.buf[:0], s...)
	_, err = e.w.Write(e.buf)
	return err
}

func (e *Encoder) WriteBytes(b []byte) error {
	_, err := e.w.Write(b)
	return err
}

func (e *Encoder) WriteByte(b byte) error {
	if len(e.buf) == 0 {
		e.buf = append(e.buf, b)
	} else {
		e.buf[0] = b
	}
	_, err := e.w.Write(e.buf[:1])
	return err
}

type Decoder struct {
	r   io.Reader
	buf []byte
}

func (d *Decoder) Init(r io.Reader) *Decoder {
	d.r = r
	d.buf = nil
	return d
}

func (d *Decoder) ReadNum() (uint64, error) {
	if len(d.buf) == 0 {
		d.buf = append(d.buf, 0)
	}
	n, s := uint64(0), uint(0)
	_, err := d.r.Read(d.buf[:1])
	b := d.buf[0]
	for err == nil {
		n = n | uint64(b&0x7f)<<s
		if (b & 0x80) == 0 {
			return n, nil
		}
		s += 7
		_, err = d.r.Read(d.buf[:1])
		b = d.buf[0]
	}
	return 0, err

}

func (d *Decoder) ReadString() (string, error) {
	n, err := d.ReadNum()
	if n == 0 || err != nil {
		return "", err
	}
	d.buf = d.buf[:0]
	for i := uint64(0); i < n; i++ {
		d.buf = append(d.buf, 0)
	}
	for i := 0; n > 0; {
		nn, err := d.r.Read(d.buf[i:])
		if err != nil {
			return "", err
		}
		n -= uint64(nn)
		i += nn
	}
	return string(d.buf), nil
}

func (d *Decoder) ReadBytes(b []byte) error {
	n := len(b)
	for i := 0; n > 0; {
		nn, err := d.r.Read(b[i:])
		if err != nil {
			return err
		}
		n -= nn
		i += nn
	}
	return nil
}

func (d *Decoder) ReadByte() (byte, error) {
	if len(d.buf) == 0 {
		d.buf = make([]byte, 1)
	}
	_, err := d.r.Read(d.buf[:1])
	return d.buf[0], err
}
