package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// подготовка
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err, "Ошибка подключения к базе данных")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// добавление
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err, "Ошибка добавления посылки")
	require.NotEmpty(t, parcel.Number, "ID посылки не должен быть пустым")

	// получение
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel

	getParcel, err := store.Get(parcel.Number)
	require.NoError(t, err, "Ошибка получения посылки")
	assert.Equal(t, parcel, getParcel, "Полученная посылка должна совпадать с добавленной")

	// удаление
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД

	err = store.Delete(parcel.Number)
	require.NoError(t, err, "Ошибка удаления посылки")

	_, err = store.Get(parcel.Number)
	require.ErrorIs(t, sql.ErrNoRows, err, "Ошибка должна указывать на отсутствие строки")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// подготовка
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err, "Ошибка подключения к базе данных")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// добавление
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка добавления посылки")
	assert.NotEmpty(t, id, "ID посылки не должен быть пустым")

	// обновление адреса
	// обновите адрес, убедитесь в отсутствии ошибки

	newAddress := "Новый тестовый адрес"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "Ошибка обновления адреса")

	// проверка
	// получите добавленную посылку и убедитесь, что адрес обновился

	getParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка получения посылки")
	assert.Equal(t, newAddress, getParcel.Address, "Адрес посылки должен быть обновлен")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// подготовка

	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err, "Ошибка подключения к базе данных")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// добавление
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка добавления посылки")
	assert.NotEmpty(t, id, "ID посылки не должен быть пустым")

	// обновление статуса
	// обновите статус, убедитесь в отсутствии ошибки

	err = store.SetStatus(id, ParcelStatusDelivered)
	require.NoError(t, err, "Ошибка обновления статуса")

	// проверка
	// получите добавленную посылку и убедитесь, что статус обновился

	getParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка получения посылки")
	assert.Equal(t, ParcelStatusDelivered, getParcel.Status, "Статус посылки должен быть обновлен")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// подготовка
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err, "Ошибка подключения к базе данных")
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// добавление
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "Ошибка добавления посылки")
		assert.NotEmpty(t, id) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

		// обновляем идентификатор добавленной у посылки

		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки

		parcelMap[id] = parcels[i]
	}

	// получение по клиенту
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "Ошибка получения посылок по клиенту")
	assert.Len(t, parcels, len(storedParcels), "Количество полученных посылок должно совпадать с добавленными")

	// проверка
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, parcelMap[parcel.Number], parcel, "Полученная посылка должна совпадать с добавленной")
	}
}
