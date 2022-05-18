// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package pagination

type Page struct {
	Number int64
	Size   int64
}

func (p *Page) Skip() int64 {
	if p.Number <= 0 || p.Size <= 0 {
		return 0
	}
	return (p.Number - 1) * p.Size
}
