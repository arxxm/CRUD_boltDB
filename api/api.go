package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/boltdb/bolt"
)

type Bucket string

const (
	ProductsBucket Bucket = "products_bucket"
	IndexBucket    Bucket = "index_bucket"
)

type Repository struct {
	db *bolt.DB
}

type Product struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type Answer struct {
	Status    string `json:"status"`
	ProductId int
	// Er        string
}

func NewRepository(db *bolt.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AddProduct(p Product) error {

	var idG int

	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ProductsBucket))
		id, _ := b.NextSequence()
		p.Id = int(id)
		idG = p.Id
		return b.Put([]byte(IntToByte(p.Id)), Encode(&p))
	})
	if err != nil {
		return err
	}

	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(IndexBucket))
		return b.Put([]byte(p.Name), IntToByte(idG))
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) ShowProducts() ([]Product, error) {
	var s []Product

	err := r.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(ProductsBucket))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			s = append(s, *Decode(v))
		}

		return nil
	})

	if err != nil {
		return []Product{}, err
	}

	return s, nil
}

func (r *Repository) DeleteAll() error {

	//Удалить из бакета с продуктами все данные key:id val:Product
	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ProductsBucket))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			err := b.Delete(k)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	//удалить из бакета с индексами все данные key:name val:id
	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(IndexBucket))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			err := b.Delete(k)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) ShowIndex() ([]string, error) {

	var sl []string

	err := r.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(IndexBucket))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("index bucket: key=%s, value=%s\n", k, v)
			sl = append(sl, string(k))
		}

		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return sl, nil
}

func (r *Repository) GetAllProducts() []byte {

	s := []Product{}

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ProductsBucket))
		b.ForEach(func(k, v []byte) error {
			s = append(s, *Decode(v))
			return nil
		})
		return nil
	})
	if err != nil {
		return nil
	}

	b, err := json.Marshal(s)
	if err != nil {
		return nil
	}
	return b
}

func (r *Repository) GetAllProductsIndexBucket() *[]string {

	var s []string

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(IndexBucket))
		b.ForEach(func(k, v []byte) error {
			str := string(k)
			s = append(s, str)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil
	}
	return &s
}

// func createAnswer(status string, id int) []byte {

// 	fail := Answer{
// 		Status: status,
// 		// Er:        er,
// 		ProductId: id,
// 	}

// 	f, _ := json.Marshal(fail)
// 	return f
// }

func (r *Repository) SearchByName(name string) (*Product, error) {

	var id int
	P := &Product{}

	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(IndexBucket))
		idByte := b.Get([]byte(name))
		idLocal, err := strconv.Atoi(string(idByte))
		id = idLocal
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ProductsBucket))
		data := b.Get([]byte(IntToByte(id)))
		P = Decode(data)
		return nil

	})
	if err != nil {
		return &Product{}, err
	}

	return P, nil
}

func (r *Repository) CheckName(name string) (error, bool) {

	var ans bool

	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(IndexBucket))
		id := b.Get([]byte(name))
		if id != nil {
			ans = true
		} else {
			ans = false
		}
		return nil
	})
	if err != nil {
		return err, false
	}

	return nil, ans
}

func (r *Repository) GetProduct(bucket string, id int) (*Product, error) {

	P := &Product{}

	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		data := b.Get([]byte(IntToByte(id)))
		P = Decode(data)
		return nil

	})
	if err != nil {
		return &Product{}, err
	}
	if P.Name == "" {
		return &Product{}, errors.New("product not found")
	}

	return P, nil
}

func (r *Repository) EditProduct(newName string, newPrice, id int) error {

	var oldName string

	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ProductsBucket))
		data := b.Get(IntToByte(id))
		P := Decode(data)
		if P.Name == "" {
			return errors.New("product not found")
		}
		oldName = P.Name
		P.Name = newName
		P.Price = newPrice
		return b.Put(IntToByte(id), Encode(P))
	})
	if err != nil {
		return err
	}

	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(IndexBucket))
		err := b.Delete([]byte(oldName))
		if err != nil {
			return err
		}
		return b.Put([]byte(newName), IntToByte(id))
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteProduct(id int) error {

	var name string

	//Удалить из бакета с продуктами по индексу key:id val:Product
	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ProductsBucket))
		product := Decode(b.Get(IntToByte(id)))
		name = product.Name
		idB := IntToByte(id)
		return b.Delete([]byte(idB))
	})
	if err != nil {
		return err
	}

	//удалить из бакета с индексами по Наименованию key:name val:id
	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(IndexBucket))
		return b.Delete([]byte(name))
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteProductByIndex(id int) error {

	var name string

	//Удалить из бакета с продуктами по индексу key:id val:Product
	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ProductsBucket))
		product := Decode(b.Get(IntToByte(id)))
		name = product.Name
		// idB := IntToByte(id)
		// return b.Delete([]byte(idB))
		return nil
	})
	if err != nil {
		return err
	}

	//удалить из бакета с индексами по Наименованию key:name val:id
	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(IndexBucket))
		return b.Delete([]byte(name))
	})
	if err != nil {
		return err
	}

	return nil
}
