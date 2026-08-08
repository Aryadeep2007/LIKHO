package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Ek post. Bas itna hi. Koi ORM, koi migration, koi schema file nahi.
type Post struct {
	Slug      string
	Title     string
	Body      string // raw markdown, jaisa user ne likha
	Tags      []string
	Created   time.Time
	Updated   time.Time
	Published bool
	Views     int

	// Reading analytics ke liye. Post ko 10 hisso me baanta hai,
	// Depth[3] matlab kitne log 40% tak pahunche. Isse pata chalta hai
	// log kahan bore hoke bhaag jate hain.
	Depth [10]int
}

// Revision file ka naam isi format me banta hai. Sort bhi isi se hota hai
// (dictionary order = time order, isliye ye format chuna hai).
const revLayout = "20060102-150405.000"

// Purana version. Har save pe pichhla wala yahan chala jata hai.
type Revision struct {
	Stamp time.Time
	Title string
	Body  string
	File  string
}

type Store struct {
	mu    sync.RWMutex
	dir   string
	posts map[string]*Post
}

func OpenStore(dir string) (*Store, error) {
	// dono folder bana lo pehle hi, warna baad me save ke waqt fail hota hai
	for _, sub := range []string{"posts", "revisions"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return nil, err
		}
	}

	s := &Store{dir: dir, posts: map[string]*Post{}}

	files, err := filepath.Glob(filepath.Join(dir, "posts", "*.md"))
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		p, err := readPostFile(f)
		if err != nil {
			// ek kharab file poore blog ko na rok de. Skip karke aage badho.
			fmt.Printf("  ⚠️  %s padha nahi gaya (%v) - skip kar rahe hain\n", filepath.Base(f), err)
			continue
		}
		s.posts[p.Slug] = p
	}
	return s, nil
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.posts)
}

// Saare posts, naye pehle. Har baar sort karte hain - 50 posts ke liye
// ye bilkul theek hai, isko optimize karna waste of time hoga.
func (s *Store) All() []*Post {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*Post, 0, len(s.posts))
	for _, p := range s.posts {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Created.After(out[j].Created)
	})
	return out
}

func (s *Store) Published() []*Post {
	var out []*Post
	for _, p := range s.All() {
		if p.Published {
			out = append(out, p)
		}
	}
	return out
}

func (s *Store) Get(slug string) (*Post, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.posts[slug]
	return p, ok
}

// Editor se save. Yahan pichhla version revisions/ me copy hota hai
// taki "time travel" feature kaam kare.
func (s *Store) Save(p *Post) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	old, existed := s.posts[p.Slug]
	if existed {
		// purana content revision bana do. Agar ye fail bhi ho jaye toh
		// save nahi rokna chahiye - history nice-to-have hai, post zaroori hai.
		if err := s.snapshot(old); err != nil {
			fmt.Printf("  ⚠️  history save nahi hui: %v\n", err)
		}
		// views aur depth editor se nahi aate, purane wale hi rakho
		p.Views = old.Views
		p.Depth = old.Depth
		p.Created = old.Created
	}
	if p.Created.IsZero() {
		p.Created = time.Now()
	}
	p.Updated = time.Now()

	if err := writePostFile(s.path(p.Slug), p); err != nil {
		return err
	}
	s.posts[p.Slug] = p
	return nil
}

func (s *Store) Delete(slug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.posts[slug]; !ok {
		return fmt.Errorf("post nahi mila: %s", slug)
	}
	if err := os.Remove(s.path(slug)); err != nil && !os.IsNotExist(err) {
		return err
	}
	delete(s.posts, slug)
	return nil
}

// View count badhao. Har view pe file likhna technically fizool hai,
// par isse count restart ke baad bhi bacha rehta hai - aur hackathon
// demo me "views: 0" dikhna sabse bada mood-off hai.
func (s *Store) AddView(slug string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.posts[slug]
	if !ok {
		return
	}
	p.Views++
	_ = writePostFile(s.path(p.Slug), p)
}

// Reader kitna neeche tak scroll kiya, wo record karo (0 se 9 tak).
func (s *Store) RecordDepth(slug string, decile int) {
	if decile < 0 || decile > 9 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.posts[slug]
	if !ok {
		return
	}
	// jitna neeche gaya, uske pehle ke saare hisse bhi padhe hi honge
	for i := 0; i <= decile; i++ {
		p.Depth[i]++
	}
	_ = writePostFile(s.path(p.Slug), p)
}

func (s *Store) snapshot(p *Post) error {
	dir := filepath.Join(s.dir, "revisions", p.Slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// milliseconds bhi daal rahe hain. Pehle sirf seconds tak tha, par agar
	// koi jaldi jaldi do baar Ctrl+S dabaye toh dono ka naam same ban jata tha
	// aur pehla version chupchap overwrite ho jata tha. Ab aisa nahi hoga.
	name := time.Now().Format(revLayout) + ".md"
	return writePostFile(filepath.Join(dir, name), p)
}

// Ek post ki saari purani versions, latest pehle.
func (s *Store) Revisions(slug string) []Revision {
	files, _ := filepath.Glob(filepath.Join(s.dir, "revisions", slug, "*.md"))

	var out []Revision
	for _, f := range files {
		p, err := readPostFile(f)
		if err != nil {
			continue
		}
		stamp, err := time.ParseInLocation(revLayout,
			strings.TrimSuffix(filepath.Base(f), ".md"), time.Local)
		if err != nil {
			stamp = p.Updated
		}
		out = append(out, Revision{
			Stamp: stamp,
			Title: p.Title,
			Body:  p.Body,
			File:  filepath.Base(f),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Stamp.After(out[j].Stamp) })
	return out
}

// Purane version pe wapas jao. Note: ye khud bhi ek naya revision banata hai,
// toh restore ko undo bhi kiya ja sakta hai. Chhota detail, par bachata hai.
func (s *Store) RestoreRevision(slug, file string) error {
	// file name sanitize - koi "../../etc/passwd" type shararat na kare
	file = filepath.Base(file)
	full := filepath.Join(s.dir, "revisions", slug, file)

	old, err := readPostFile(full)
	if err != nil {
		return fmt.Errorf("wo version padha nahi gaya: %w", err)
	}

	cur, ok := s.Get(slug)
	if !ok {
		return fmt.Errorf("post ab exist hi nahi karta: %s", slug)
	}

	cur.Title = old.Title
	cur.Body = old.Body
	cur.Tags = old.Tags
	return s.Save(cur)
}

func (s *Store) path(slug string) string {
	return filepath.Join(s.dir, "posts", slug+".md")
}

func (s *Store) DataDir() string { return s.dir }

// ---------- file padhna / likhna ----------

// Format kuch aisa hai:
//
//	---
//	title: Mera pehla post
//	tags: go, hackathon
//	---
//	yahan se markdown shuru
//
// YAML library import kar sakte the, par sirf 7 fields ke liye
// ek poori dependency laana overkill laga.
func readPostFile(path string) (*Post, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := &Post{
		Slug:      strings.TrimSuffix(filepath.Base(path), ".md"),
		Published: true,
	}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024) // lambi lines ke liye

	inHeader := false
	first := true
	var body []string

	for sc.Scan() {
		line := sc.Text()

		if first {
			first = false
			if strings.TrimSpace(line) == "---" {
				inHeader = true
				continue
			}
			// header hai hi nahi - poori file body maan lo
		}

		if inHeader {
			if strings.TrimSpace(line) == "---" {
				inHeader = false
				continue
			}
			key, val, found := strings.Cut(line, ":")
			if !found {
				continue
			}
			applyHeader(p, strings.TrimSpace(key), strings.TrimSpace(val))
			continue
		}

		body = append(body, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	p.Body = strings.TrimLeft(strings.Join(body, "\n"), "\n")
	if p.Title == "" {
		p.Title = p.Slug
	}
	if p.Created.IsZero() {
		// file ki date use kar lo, kuch na kuch toh dikhna chahiye
		if st, err := os.Stat(path); err == nil {
			p.Created = st.ModTime()
		} else {
			p.Created = time.Now()
		}
	}
	if p.Updated.IsZero() {
		p.Updated = p.Created
	}
	return p, nil
}

func applyHeader(p *Post, key, val string) {
	switch strings.ToLower(key) {
	case "title":
		p.Title = val
	case "tags":
		p.Tags = splitTags(val)
	case "created":
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			p.Created = t
		}
	case "updated":
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			p.Updated = t
		}
	case "published":
		p.Published = val != "false" && val != "no" && val != "0"
	case "views":
		p.Views, _ = strconv.Atoi(val)
	case "depth":
		for i, part := range strings.Split(val, ",") {
			if i >= 10 {
				break
			}
			p.Depth[i], _ = strconv.Atoi(strings.TrimSpace(part))
		}
	}
}

func writePostFile(path string, p *Post) error {
	var b strings.Builder

	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %s\n", p.Title)
	fmt.Fprintf(&b, "tags: %s\n", strings.Join(p.Tags, ", "))
	fmt.Fprintf(&b, "created: %s\n", p.Created.Format(time.RFC3339))
	fmt.Fprintf(&b, "updated: %s\n", p.Updated.Format(time.RFC3339))
	fmt.Fprintf(&b, "published: %t\n", p.Published)
	fmt.Fprintf(&b, "views: %d\n", p.Views)

	nums := make([]string, 10)
	for i, d := range p.Depth {
		nums[i] = strconv.Itoa(d)
	}
	fmt.Fprintf(&b, "depth: %s\n", strings.Join(nums, ","))
	b.WriteString("---\n\n")
	b.WriteString(p.Body)
	b.WriteString("\n")

	// pehle temp file me likho, phir rename. Agar beech me power chali jaye
	// toh aadhi likhi file se post corrupt nahi hoga.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ---------- chhote helpers ----------

var slugKill = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugKill.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		// title agar poora hindi/emoji me hai toh slug khali reh jayega,
		// aise me timestamp se kaam chala lete hain
		s = "post-" + time.Now().Format("20060102-150405")
	}
	if len(s) > 60 {
		s = strings.Trim(s[:60], "-")
	}
	return s
}

func splitTags(s string) []string {
	var out []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func readingTime(body string) int {
	// average banda ~200 word/min padhta hai. Kam se kam 1 min dikhao,
	// "0 min read" ajeeb lagta hai.
	n := len(strings.Fields(body))
	m := n / 200
	if m < 1 {
		return 1
	}
	return m
}

func humanTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "abhi abhi"
	case d < time.Hour:
		return fmt.Sprintf("%d min pehle", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d ghante pehle", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%d din pehle", int(d.Hours()/24))
	default:
		return t.Format("2 Jan 2006")
	}
}
