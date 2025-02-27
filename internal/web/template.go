package web

import (
	"log/slog"
	"net/http"
	"strings"
)

type Template interface {
	GetTemplatesLocation() string
	GetStaticLocation() string
	GetLayoutTpl() string
	GetDefaultContentTpl() string
	GetData(r *http.Request, contentData map[string]any) (map[string]any, error)
}

type DefaultTemplate struct {
	code     string
	log      *slog.Logger
	menuData []MenuItem
}

type MenuItem struct {
	Title    string
	Link     string
	IsActive bool
	Children []MenuItem
}

func NewDefaultTemplate(log *slog.Logger, menuData []MenuItem) *DefaultTemplate {
	return &DefaultTemplate{
		code:     "default",
		log:      log,
		menuData: menuData,
	}
}

func (t *DefaultTemplate) GetTemplatesLocation() string {
	return "templates/" + t.code
}

func (t *DefaultTemplate) GetStaticLocation() string {
	return t.GetTemplatesLocation() + "/static"
}

func (t *DefaultTemplate) GetLayoutTpl() string {
	return "layout.html"
}

func (t *DefaultTemplate) GetDefaultContentTpl() string {
	return "content/default.html"
}

func (t *DefaultTemplate) getDefaultTitle() string {
	return "tgtag"
}

func (t *DefaultTemplate) GetData(r *http.Request, contentData map[string]any) (map[string]any, error) {
	commonData := make(map[string]any)
	commonData["Menu"] = t.getMenu(r.URL.Path)
	t.shallowMapMerge(commonData, contentData)
	return commonData, nil
}

func (t *DefaultTemplate) shallowMapMerge(map1, map2 map[string]any) {
	for k, v := range map2 {
		map1[k] = v
	}
}

func (t *DefaultTemplate) getMenu(current string) []MenuItem {
	if len(t.menuData) == 0 {
		return nil
	}
	//var result []MenuItem
	for _, mi := range t.menuData {
		mi.IsActive = t.isMenuLinkCurrent(mi.Link, current)
		if len(mi.Children) > 0 {
			for _, mimi := range mi.Children {
				mimi.IsActive = t.isMenuLinkCurrent(mimi.Link, current)
			}
		}
		//result = append(result, mi)
	}
	//t.menuData = result
	return t.menuData
}

func (t *DefaultTemplate) isMenuLinkCurrent(current, link string) bool {
	if link == "/" {
		return current == "/"
	}
	return strings.HasPrefix(current, link)
}
