package internal

import "sort"

const tagKey = "tags"
const MaxTagLength = 32

type TagsCache struct {
	*cacheStore
	Tagmap map[string]struct{}
}

func LoadTagsCache(db *db) (*TagsCache, error) {
	tags := new(TagsCache)
	tags.key = cacheKey(tagKey)
	err := tags.load(tags)
	if tags.Tagmap == nil {
		tags.Tagmap = make(map[string]struct{})
	}
	return tags, err
}

func (t *TagsCache) IsTag(tag string) bool {
	_, ok := t.Tagmap[tag]
	return ok
}

func (t *TagsCache) AddTag(tag string) error {
	if t.IsTag(tag) {
		return nil
	}
	t.Tagmap[tag] = struct{}{}
	return t.save(t)
}

//Tags returns a lexicographically sorted slice of tag strings
func (t *TagsCache) Tags() []string {
	tags := make([]string, len(t.Tagmap))
	for key, _ := range t.Tagmap {
		tags = append(tags, key)
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i] < tags[j] })
	return tags
}
