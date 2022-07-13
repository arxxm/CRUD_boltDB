package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arxxm/CRUD_test/api"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

type APIHandler struct {
	repo *api.Repository
}

var prods = []api.Product{}

func (h *APIHandler) productsPage(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "Products page")

	prods = []api.Product{}
	b := h.repo.GetAllProducts()
	if b == nil {
		fmt.Fprintf(w, "Ошибка")
	}
	err := json.Unmarshal(b, &prods)
	if err != nil {
		log.Fatal(err)
	}

	t, _ := template.ParseFiles("templates/products.html", "templates/header.html", "templates/footer.html")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	t.ExecuteTemplate(w, "products", prods)
}

func (h *APIHandler) check(w http.ResponseWriter, r *http.Request) {

	// var names []string

	names, _ := h.repo.ShowIndex()
	// for _, v := range names {
	// 	fmt.Println(v)
	// }
	// if err != nil {
	// 	log.Fatal(err)
	// }

	t, _ := template.ParseFiles("templates/check.html", "templates/header.html", "templates/footer.html")
	t.ExecuteTemplate(w, "check", names)
}

func NewAPIHandler(repo *api.Repository) (*APIHandler, error) {
	var h = APIHandler{}
	h.repo = repo

	return &h, nil
}

// func (h *APIHandler) saveProduct(w http.ResponseWriter, r *http.Request) {

// 	// r.Form.Get()
// 	newProduct := api.Product{
// 		Id:    123,
// 		Name:  "Phone",
// 		Price: 15000,
// 	}

// 	err := h.repo.AddProduct(string(api.ProductsBucket), &newProduct)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Fprintf(w, "Продукт успешно записан")
// }

func (h *APIHandler) getProduct(w http.ResponseWriter, r *http.Request) {

	P, err := h.repo.GetProduct(string(api.ProductsBucket), 123)
	if err != nil {
		log.Fatal(err)
	}

	_, err = fmt.Fprintf(w, P.Name)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *APIHandler) editProductById(w http.ResponseWriter, r *http.Request) {

	// err := h.repo.EditProduct(string(api.ProductsBucket), "iPhone XR", 123, 15000)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Fprintf(w, "Продукт успешно изменен")
	vars := mux.Vars(r)
	t, err := template.ParseFiles("templates/edit.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Fatal(err)
	}

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Fatal(err)
	}

	P, err := h.repo.GetProduct(string(api.ProductsBucket), id)
	if err != nil {
		log.Fatal(err)
	}

	t.ExecuteTemplate(w, "edit", P)
}

func (h *APIHandler) editProduct(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := r.FormValue("name")
	priceStr := r.FormValue("price")
	// fmt.Printf("vars id is: %s\n", vars["id"])
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Fatal(err)
	}

	price, err := strconv.Atoi(priceStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("name is: %s, price: %d, id: %d\n", name, price, id)
	err = h.repo.EditProduct(name, price, id)
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, "/products/", http.StatusSeeOther)

}

func (h *APIHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Fatal(err)
	}
	t.ExecuteTemplate(w, "create", nil)
}

func (h *APIHandler) searchByName(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("name")
	fmt.Printf("Имя: %s\n", name)
	p, err := h.repo.SearchByName(name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Найденный товар: %s, цена: %d\n", p.Name, p.Price)

	prods = []api.Product{}
	prods = append(prods, *p)
	t, _ := template.ParseFiles("templates/found.html", "templates/header.html", "templates/footer.html")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	t.ExecuteTemplate(w, "found", prods)

}

func (h *APIHandler) addProduct(w http.ResponseWriter, r *http.Request) {

	// var s []string
	name := r.FormValue("name")
	priceStr := r.FormValue("price")
	// s = append(s, name)
	// s = append(s, priceStr)

	// t, _ := template.ParseFiles("templates/check.html", "templates/header.html", "templates/footer.html")
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
	// t.ExecuteTemplate(w, "check", s)

	if name == "" || priceStr == "" {
		fmt.Fprintf(w, "Все поля должны быть заполнены")
	}

	price, err := strconv.Atoi(priceStr)
	if err != nil {
		log.Fatal(err)
	}

	p := api.Product{
		Name:  name,
		Price: price,
	}

	// b, err := json.Marshal(p)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	err, ans := h.repo.CheckName(name)
	if err != nil {
		log.Fatal(err)
	}
	if ans {
		fmt.Fprintf(w, "Продукт с таким именем уже занесен в базу")
		fmt.Println("Продукт с таким именем уже занесен в базу")
		return
	}
	// var ans api.Answer
	err = h.repo.AddProduct(p)
	if err != nil {
		log.Fatal(err)
	}
	// err = json.Unmarshal(a, &ans)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// if ans.Status == "failure" {
	// 	fmt.Fprintf(w, "Не удалось добавить продукт")
	// 	return
	// }

	http.Redirect(w, r, "/products/", http.StatusSeeOther)

}

func (h *APIHandler) deleteAll(w http.ResponseWriter, r *http.Request) {
	h.repo.DeleteAll()
	http.Redirect(w, r, "/products/", http.StatusSeeOther)
}

func (h *APIHandler) deleteProduct(w http.ResponseWriter, r *http.Request) {

	// var s []int
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Fatal(err)
	}

	// name := r.FormValue("name")

	// s = append(s, id)

	// t, _ := template.ParseFiles("templates/check.html", "templates/header.html", "templates/footer.html")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// t.ExecuteTemplate(w, "check", s)
	err = h.repo.DeleteProduct(id)
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, "/products/", http.StatusSeeOther)
}
func (h *APIHandler) deleteProductByIndex(w http.ResponseWriter, r *http.Request) {

	// var s []int
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Fatal(err)
	}

	// name := r.FormValue("name")

	// s = append(s, id)

	// t, _ := template.ParseFiles("templates/check.html", "templates/header.html", "templates/footer.html")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// t.ExecuteTemplate(w, "check", s)
	err = h.repo.DeleteProductByIndex(id)
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, "/check/", http.StatusSeeOther)

}

// func (h *APIHandler) ShowKeys(w http.ResponseWriter, r *http.Request) {
// 	err := h.repo.ShowAllKeys(string(api.ProductsBucket))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

func (h *APIHandler) showPr(w http.ResponseWriter, r *http.Request) {
	_, err := h.repo.ShowProducts()
	if err != nil {
		log.Fatal(err)
	}

}

func (h *APIHandler) showIn(w http.ResponseWriter, r *http.Request) {
	_, err := h.repo.ShowIndex()
	if err != nil {
		log.Fatal(err)
	}

}

func (h *APIHandler) serveHome(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("templates/home.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Fatal(err)
	}
	t.ExecuteTemplate(w, "home", nil)
}

// func (h *APIHandler) Posts(w http.ResponseWriter, r *http.Request) { /* … */ }

func main() {

	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	repo := api.NewRepository(db)

	h, err := NewAPIHandler(repo)
	if err != nil {
		log.Fatal(err)
	}

	m := http.NewServeMux()
	rtr := mux.NewRouter()

	rtr.HandleFunc("/", h.serveHome).Methods("GET")
	// rtr.HandleFunc("/new/", h.saveProduct).Methods("GET")
	rtr.HandleFunc("/products/", h.productsPage).Methods("GET")
	rtr.HandleFunc("/get/", h.getProduct).Methods("GET")
	rtr.HandleFunc("/product/edit/{id}", h.editProductById).Methods("GET")
	rtr.HandleFunc("/cmd/delete-product/{id}", h.deleteProduct)
	rtr.HandleFunc("/cmd/delete-product-by-index/{id}", h.deleteProductByIndex)
	rtr.HandleFunc("/cmd/edit-product/{id}", h.editProduct)
	rtr.HandleFunc("/product/add", h.createProduct)
	rtr.HandleFunc("/cmd/add-product", h.addProduct)
	rtr.HandleFunc("/showpr", h.showPr)
	rtr.HandleFunc("/showind", h.showIn)
	rtr.HandleFunc("/check", h.check)
	rtr.HandleFunc("/cmd/delete-all", h.deleteAll)
	rtr.HandleFunc("/q/product-search-by-name", h.searchByName)

	m.Handle("/", rtr)
	var srv = &http.Server{
		Addr:    ":8080",
		Handler: m,
		// TLSConfig: nil,

		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    16 * 1024,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal()
	}
}

func initDB() (*bolt.DB, error) {

	db, err := bolt.Open("crud.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	// defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(api.ProductsBucket))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(api.IndexBucket))
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return db, nil
}
