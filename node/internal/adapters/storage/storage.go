package storage

import (
	"fmt"
	"log"
	"node/internal/block"

	"go.etcd.io/bbolt"
)

const (
	blocksBucketName = "blocks_bucket"
)

/* Структура драйвера для базы данных */
type BBoltDBDriver struct {
	*bbolt.DB
}

/* Конструктор */
func NewBBoltDBDriver() *BBoltDBDriver {
	var driver BBoltDBDriver

	db, err := bbolt.Open("../bolt.db", 0600, nil)
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
WriteValue Функция записи значения по
ключу в определенную корзину

	bucketName имя корзины
	key ключ
	value данные
*/
func (driver *BBoltDBDriver) WriteValue(bucketName, key, value []byte) error {
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
ReadValue Функция чтения значения по ключу из
определенной корзину

	bucketName имя корзины
	key ключ
*/
func (driver *BBoltDBDriver) ReadValue(bucketName, key []byte) ([]byte, error) {
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

/*
IsBlockchainExist Функция для удовлетворения
интерфейсу BlockchainStorage. Проверяет существует
ли блокчейн в файловой системе
*/
func (driver *BBoltDBDriver) IsBlockchainExist() (bool, error) {
	err := driver.DB.View(func(tx *bbolt.Tx) error {
		// Получаем корзину (bucket) с именем "MyBucket"
		bucket := tx.Bucket([]byte(blocksBucketName))
		if bucket != nil {
			return fmt.Errorf("")
		}

		return nil
	})

	return err != nil, nil
}

/*
MakeNewBlockchain Функция для удовлетворения
интерфейсу BlockchainStorage. Проводит создание
генезис блока и его загрузку на диск
*/
func (driver *BBoltDBDriver) MakeNewBlockchain(genesisBlock *block.Block) error {
	err := driver.DB.Update(func(tx *bbolt.Tx) error {
		// Создаем корзину (bucket)
		bucket, err := tx.CreateBucketIfNotExists([]byte(blocksBucketName))
		if err != nil {
			return err
		}

		// Сериализуем блок
		data, err := genesisBlock.SerializeBlock()
		if err != nil {
			return fmt.Errorf("Can not write genesis block: %v", err)
		}

		// Кладем блок в корзину
		err = bucket.Put(genesisBlock.Hash, data)
		if err != nil {
			return fmt.Errorf("Can not write genesis block: %v", err)
		}

		// Кладем хеш блока в корзину, создавая тем самым хвостик
		err = bucket.Put([]byte("l"), genesisBlock.Hash)
		if err != nil {
			return fmt.Errorf("Can not write hash from block: %v", err)
		}

		return nil
	})

	return err
}

/*
BlockchainGetTip Функция для удовлетворения
интерфейсу BlockchainStorage. Проводит создание
генезис блока и его загрузку на диск
*/
func (driver *BBoltDBDriver) BlockchainGetTip() ([]byte, error) {
	tip, err := driver.ReadValue([]byte(blocksBucketName), []byte("l"))
	return tip, err
}

/*
WriteNewBlock Функция для удовлетворения
интерфейсу BlockchainStorage. Сохраняет на диск
новый блок в блокчейн и обновляет хвостик
*/
func (driver *BBoltDBDriver) WriteNewBlock(newBlock *block.Block, lastHash []byte) error {
	err := driver.DB.Update(func(tx *bbolt.Tx) error {
		// Читаем корзину
		bucket, err := tx.CreateBucketIfNotExists([]byte(blocksBucketName))
		if err != nil {
			return fmt.Errorf("Bucket not found")
		}

		// Сериализуем блок
		data, err := newBlock.SerializeBlock()
		if err != nil {
			return fmt.Errorf("Can not write block: %v", err)
		}

		// Кладем блок в корзину
		err = bucket.Put(newBlock.Hash, data)
		if err != nil {
			return fmt.Errorf("Can not write block: %v", err)
		}

		// Кладем хеш блока в корзину, создавая тем самым хвостик
		err = bucket.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			return fmt.Errorf("Can not write hash from block: %v", err)
		}

		return nil
	})

	return err
}

/*
GetExistBlockByHash Функция для удовлетворения
интерфейсу BlockchainStorage. Вычитывает из базы данныз блок
и парсит его в используемую структуру
*/
func (driver *BBoltDBDriver) GetExistBlockByHash(hash []byte) (*block.Block, error) {
	data, err := driver.ReadValue([]byte(blocksBucketName), hash)
	if err != nil {
		return nil, fmt.Errorf("Can not get block by hash: %v", err)
	}

	b, err := block.DeserializeBlock(data)
	if err != nil {
		return nil, fmt.Errorf("Read block was failed with convert error: %v", err)
	}

	return b, nil
}

/* 
Мини-итератор, пригодится для отладки,
выводит всю базу данных целиком 
*/
func (driver *BBoltDBDriver) Iterator() {
	driver.DB.View(func(tx *bbolt.Tx) error {
		tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			log.Printf("Корзина: %v\n", name)
			b.ForEach(func(k, v []byte) error {
				log.Printf("  Ключь: %v, Значение: %v\n", k, v)
				return nil
			})
			return nil
		})
		return nil
	})
}
