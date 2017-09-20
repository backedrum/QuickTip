/*
ViewLater Chrome Extension
Copyright (C) 2017 Andrii Zablodskyi (andrey.zablodskiy@gmail.com)

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/
package main

import (
	"encoding/json"
	"github.com/fabioberger/chrome"
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
	"strconv"
)

type Link struct {
	Id         int    `json:"id"`
	Url        string `json:"url"`
	Title      string `json:"title"`
	Screenshot string `json:"screenshot"`
}

type LinksList []Link

var storage = js.Global.Get("localStorage")

var urls = LinksList{}

var screenshot = ""

func main() {

	c := chrome.NewChrome()
	d := dom.GetWindow().Document()

	table := d.GetElementByID("links")

	// get json from storage
	linksJson := storage.Get("readLaterLinks").String()

	if linksJson != "" {
		err := json.Unmarshal([]byte(linksJson), &urls)
		if err != nil {
			println("Error has been occurred ", err.Error())
		}
	}

	for _, link := range urls {
		row := d.CreateElement("tr").(*dom.HTMLTableRowElement)
		urlColumn := d.CreateElement("td").(*dom.HTMLTableCellElement)
		addRemoveButton(d, table, row, link.Id)
		row.AppendChild(urlColumn)

		urlColumn.SetAttribute("id", strconv.Itoa(link.Id))
		urlColumn.SetInnerHTML("<a href=\"" + link.Url + "\">" + link.Title + "</a>")

		addScreenshot(d, row, link.Screenshot)
		table.AppendChild(row)
	}

	add := d.GetElementByID("addButton").(*dom.HTMLInputElement)
	add.Call("addEventListener", "click", func(event *js.Object) {
		row := d.CreateElement("tr").(*dom.HTMLTableRowElement)

		c.Windows.GetCurrent(chrome.Object{}, func(window chrome.Window) {
			id := window.Id

			c.Tabs.CaptureVisibleTab(id, nil, func(dataUrl string) {
				screenshot = dataUrl
			})

			c.Tabs.GetSelected(id, func(tab chrome.Tab) {
				currentUrl := tab.Url

				nextId := 0
				if len(urls) > 0 {
					nextId = urls[len(urls)-1].Id + 1
				}

				addRemoveButton(d, table, row, nextId)

				urlColumn := d.CreateElement("td").(*dom.HTMLTableCellElement)
				urlColumn.SetAttribute("id", strconv.Itoa(nextId))
				urlColumn.SetInnerHTML("<a href=\"" + currentUrl + "\">" + tab.Title + "</a>")
				row.AppendChild(urlColumn)

				addScreenshot(d, row, screenshot)
				table.AppendChild(row)

				urls = append(urls, Link{Id: nextId, Url: currentUrl, Title: tab.Title, Screenshot: screenshot})
				marshalUrlsToStorage()
			})

		})
	})
}

func addScreenshot(d dom.Document, tableRow *dom.HTMLTableRowElement, screenshot string) {
	screenshotColumn := d.CreateElement("td").(*dom.HTMLTableCellElement)

	div := d.CreateElement("div").(*dom.HTMLDivElement)
	div.Style().Set("height", "90px")
	div.Style().Set("width", "240px")

	img := d.CreateElement("img").(*dom.HTMLImageElement)
	img.SetAttribute("src", screenshot)
	img.Style().Set("max-height", "100%")
	img.Style().Set("max-width", "100%")
	div.AppendChild(img)

	screenshotColumn.AppendChild(div)
	tableRow.AppendChild(screenshotColumn)
}

func addRemoveButton(d dom.Document, table dom.Element, tableRow *dom.HTMLTableRowElement, urlId int) {
	removeButtonColumn := d.CreateElement("td").(*dom.HTMLTableCellElement)

	remove := d.CreateElement("input").(*dom.HTMLInputElement)
	remove.SetAttribute("type", "button")
	remove.SetAttribute("value", "X")
	remove.SetClass("button removeButton")

	remove.Call("addEventListener", "click", func(event *js.Object) {
		removeLink(urlId)
		marshalUrlsToStorage()
		table.RemoveChild(tableRow)
	})
	removeButtonColumn.AppendChild(remove)
	tableRow.AppendChild(removeButtonColumn)
}

func removeLink(id int) {
	modified := LinksList{}
	for _, link := range urls {
		if link.Id == id {
			continue
		}

		modified = append(modified, link)
	}

	urls = modified
}

func marshalUrlsToStorage() {
	bytes, err := json.Marshal(urls)
	if err != nil {
		println("Error has occurred during marshalling ", err.Error())
	}

	storage.Set("readLaterLinks", string(bytes))
}
