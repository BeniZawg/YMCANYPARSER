package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly"
)

/*
 * Создаю сласл офисов и канал результата

 * Прохожусь по базовой странице
 * Заполняю объект офиса
 * Кладу в контекст объект офиса конкретного под именем юрл
 * Начинаю парсить инфу о офисе
 *
 * Прохожусь по странице офиса для каждого работника
 * Достаю из контекста объект офиса по юрл
 * Заполняю карту сотрудника и добавляю в слайс сотрудников офиса
 * Кладу обновленный объект офиса в контекст
 *
 * ПРОБЛЕМА1: *магически? когда спарсил всех сотрудников, кладу объект офиса в канал*
 *
 * ПРОБЛЕМА2: *магически? когда спарсил все офиса, вытаскиваю все данные из канала и закрываю его*
 *
 * ПРОБЛЕМА3: *магически? понимаю все форматирования вёрстки со всеми вариакциями наличия полей и не умираю от этого*
 *
 * Заполняю слайс офисов
 *
 * ???
 *
 * PROFIT
 */

// TODO: Передавать структуру с адресом всем дочерним визитам через контекст
// TODO: Заполять сотрудников в структуре
// TODO: подтянуть геокодер к адресу
// TODO: откомментировать всё

type Office struct {
	Name           string
	District       string
	Address        string
	Phone          string
	Geo            string
	OfficeEmployee []OfficeEmployee
}

type OfficeEmployee struct {
	FullName string
	Phone    string
	Email    string
	Position string
}

func main() {
	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains("ymcanyc.org"),
	)

	ListOfOffices := []Office{}

	OfficeCh := make(chan Office)

	c.OnHTML(".location-list-item", func(e *colly.HTMLElement) {
		LocalOffice := Office{
			Name:     e.ChildText(".card-type--branch"),
			District: e.ChildText(".field-borough"),
			Address:  e.ChildText(".field-location-direction"),
			Phone:    e.ChildText(".field-location-phone a"),
		}

		fmt.Printf("LocalOffice: %+v \n", LocalOffice)

		link := "https://ymcanyc.org" + e.ChildAttr(".branch-view-button a", "href") + "/about"

		e.Request.Ctx.Put(link, LocalOffice)

		e.Request.Visit(link)
	})

	c.OnHTML(".container.col-2c-container.d-flex .field-sb-body.block-share.field-item p", func(e *colly.HTMLElement) {
		if e.ChildText("strong") == "" {
			return
		}

		link := e.Request.AbsoluteURL("")

		LocalOffice := e.Request.Ctx.GetAny(link).(Office)

		Content, _ := e.DOM.Html()
		fmt.Printf("Office Obj: %+v \n\n", strings.Split(Content, "<br/>"))

		Employee := OfficeEmployee{
			FullName: e.ChildText("strong"),
			Email:    e.ChildText("a"),
			Phone:    e.ChildText(""),
		}

		LocalOffice.OfficeEmployee = append(LocalOffice.OfficeEmployee, Employee)

		e.Request.Ctx.Put(link, LocalOffice)

		fmt.Printf("Office Obj: %+v \n URL: %+v \n\n", LocalOffice, e.Request.AbsoluteURL(""))
	})

	c.Visit("https://ymcanyc.org/locations?type&amenities")

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL)
	})

	for {
		select {
		case office, ok := <-OfficeCh:
			if !ok {
				OfficeCh = nil
			} else {
				ListOfOffices = append(ListOfOffices, office)
			}
		}

		if OfficeCh == nil {
			break
		}
	}

	c.Wait()

	log.Printf("Scraping finished, check results:\n")
	log.Printf("%+v", ListOfOffices)
}
