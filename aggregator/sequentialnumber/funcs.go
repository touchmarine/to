package sequentialnumber

func (a aggregate) Group() group {
	return (&grouper{a: a}).group(1)
}

type group []interface{}

type grouper struct {
	a   aggregate
	pos int
}

func (g *grouper) group(base int) group {
	var gr group
	for g.pos < len(g.a) {
		p := g.a[g.pos]
		if depth := p.depth(); depth > base {
			gr = append(gr, g.group(depth))
		} else if depth == base {
			gr = append(gr, p)
			g.pos++
		} else {
			break
		}
	}
	return gr
}
