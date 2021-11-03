package config

// ShallowMerge merges src into dst but can only add or override whole objects.
// It does not support more granularity.
func ShallowMerge(dst, src *Config) *Config {
	if dst == nil {
		dst = &Config{}
	}
	if src == nil {
		return dst
	}
	for n, t := range src.Templates {
		if dst.Templates == nil {
			dst.Templates = map[string]string{}
		}
		dst.Templates[n] = t
	}
	for n, e := range src.Elements {
		if dst.Elements == nil {
			dst.Elements = Elements{}
		}
		dst.Elements[n] = e
	}
	for n, v := range src.Aggregates {
		if dst.Aggregates == nil {
			dst.Aggregates = Aggregates{}
		}
		dst.Aggregates[n] = v
	}
	return dst
}
