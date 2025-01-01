package boltdb

import (
	"log"
	"fmt"

	"go.etcd.io/bbolt"
)

/* Структура драйвера для базы данных */
type BBoltDBDriver struct {
	*bbolt.DB
}

/* Конструктор */
func NewBBoltDBDriver() *BBoltDBDriver {
	var driver BBoltDBDriver

	db, err := bbolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	driver.DB = db
	return &driver
}

/* Корректно закрывает базу данных */
func (driver *BBoltDBDriver) CloseConnection() {
	driver.DB.Close()
}

/* 
Функция записи значения по ключу в определенную корзину
	bucketName имя корзины
	key ключ
	value данные
*/
func (driver *BBoltDBDriver)WriteValue(bucketName, key, value []byte) error {
	return driver.DB.Update(func(tx *bbolt.Tx) error {
		// Создаем корзину (bucket)
		bucket, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}

		// Кладем данные в корзину
		err = bucket.Put(key, value)
		if err != nil {
			return err
		}

		return nil
	})
}

/* 
Функция чтения значения по ключу из определенной корзину
	bucketName имя корзины
	key ключ
*/
func (driver *BBoltDBDriver)ReadValue(bucketName, key []byte) ([]byte, error) {
	value := make([]byte, 0)
	err := driver.DB.View(func(tx *bbolt.Tx) error {
		// Получаем корзину (bucket) с именем "MyBucket"
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf("Bucket not found")
		}

		// Читаем данные по ключу
		value = bucket.Get(key)
		if value == nil {
			return fmt.Errorf("Value not found for key")
		}

		return nil
	})

	return value, err
}
