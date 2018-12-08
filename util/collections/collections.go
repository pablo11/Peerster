package collections

func Filter(vs []string, f func(string) bool) []string {
    vsf := make([]string, 0)
    for _, v := range vs {
        if f(v) {
            vsf = append(vsf, v)
        }
    }
    return vsf
}

func Map(vs []string, f func(string) string) []string {
    vsm := make([]string, len(vs))
    for i, v := range vs {
        vsm[i] = f(v)
    }
    return vsm
}

func MapKeys(vs map[string]string) []string {
    keys := make([]string, 0, len(vs))
    for key, _ := range vs {
        keys = append(keys, key)
    }
    return keys
}
