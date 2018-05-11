package main

import (
	"net/http"
	"log"
	"encoding/json"
)

type DBResp struct {

}

func (resp *DBResp) RowsAffected() (interface{}, error) {
	return nil, nil
}

type DB struct {

}

func (tdb *DB) Exec(query string) (*DBResp, error) {
	return nil, nil
}

func (tdb *DB) Query(string, ...interface{}) (*DBResp, error) {
	return nil, nil
}

var db = new(DB)

// 1. Найти и исправить принципиальные ошибки в коде, считаем все используемы объекты загружены:

// 1.1. Разбор
// Здесь я опишу комментарии, что я считаю не так с кодом


// Ошибка: нет указания значения возврата функции. Какой именно должен быть возврат укажу около return
// Ошибка: чаще всего также передается http.ResponseWriter или нечто похожее, иначе невозможно на данном уровне вписать ответ.
// Я не знаю что вы точно имели в виду, поэтому оставлю так, можно представить, что это функция, которая при
// которая при отсутсвии ошибки в возврате, будет просто отдавать 200 OK. Хотя, если сюда уже передается *http.Request
// в неразобранном виде, могу предположить, что и http.ResponseWriter требуется передать.
func object_input_insert(r *http.Request) (error, string) {
	// Ошибка: указывается переменная, которой присваивается значение,
	// но оно нигде не использутеся. Требуется убрать.
	var flag string
	// Ошибка: ранее переменная не была определена, поэтому требуется использовать символ `:=`
	// Ошибка: сюда не передается ничего из запроса, это не совсем корректно
	// Вопрос: а почему вообще вставляется `id` в базу данных? Можно делать кастомные id,
	// чтобы исключить возможность иттерации по базе из вне или скрыть количество записей.
	// Может быть это пользовательский `id`?
	// Я бы убрал это.
	query = "insert into test (id, text) values (1, 'text')"
	res, err := db.Exec(query)
	if err != nil {
		// Дополнение: по-хорошему стоит использовать готовые библиотеки для вывода логов, например, ...
		// Ошибка: данных об ошибке будет слишком мало, будет выведен только текст ошибки, но не будет стэка
		log.Print("Ошибка взаимодействия с БД (", err, ")!")
	} else {
		// Ошибка: переменной response не существует. Требуется убрать ее.
		// Ошибка: переменная ra никак не используется. Требуется заменить ее на `_`
		// Ошибка: не стоит перезаписывать вышенаходящуюся переменную ошибки err, иначе будет сложно отследить
		// в каком месте она перестала быть nil
		// Комментарий: я не знаю какая именно ДБ используется или какая библиотека, поскольку разные библиотеки могут отдавать разные данные
		ra, err := response.RowsAffected()
		// Ошибка: указывается не тот тип переменной + данная переменная вообще не используется, поэтому стоит ее убрать.
		flag = 3
		// Комментарий: странно, что результат выводится в ошибке, лучше помнять эти строки.
		log.Print("Результат: ", " (", err, ")")
	}

	// Ошибка: не стоит возвращать просто текст ошибки, поскольку возврат string из функции не говорит точно, что произошло.
	// Требуется возвращать именно ошибку, а вот во что ее преобразовывать должна решать функция, на которую
	// возложена данная ответственность (например, функция записи ответа в http.Response)
	// Комментарий: не помню, чтобы у струтукры Error было свойство `String`
	return err.String
}

// 1.1. Коррекция

type TestJSON struct {
	Text string
}

func object_input_insert_correct(w http.ResponseWriter,r *http.Request) error {
	// Комментарий: поскольку здесь производится `insert`, я предполагаю, что это POST запрос, у которого есть тело.
	// Если это не POST, то роутер не пропустит запрос до данного уровня, поэтому здесь не требуется валидация.
	// Комментарий: также, я предположу, что используется JSON. Чаще всего я использую библиотеку https://github.com/mailru/easyjson
	dc := json.NewDecoder(r.Body)
	var reqB TestJSON
	if err := dc.Decode(&reqB); err != nil {
		log.Print("Ошибка парсинга запроса (", err, ")!")
		return err
	}
	defer r.Body.Close()

	// Комментарий: чаще всего, в каждой библиотеке есть уже встроенная функция создания запроса с аргментами,
	// которая сделает грамотный escape и все нужные обработки
	resp, err := db.Query("insert into test (text) values ($1)", reqB.Text)

	if err != nil {
		log.Print("Ошибка взаимодействия с БД (", err, ")!")
		// Комментарий: статус и текст ошибки будут зависеть от типа самой ошибки. Если проблема с коннектом к ДБ,
		// то это 500 и проблема сервера + вызов дополнительных функций срочного оповещения админов, если неккоректные данные,
		// то это уже 422 или нечто похожее
		w.WriteHeader(http.StatusBadRequest)
		// Комментарий: это очень некорректный ответ, но пока так.
		w.Write([]byte(err.Error()))
		// Комментарий: Я предпочитаю явно возвращать ошибку, в месте, где она должна оставновить код
		return err
	} else {
		_, errRows := resp.RowsAffected()
		if errRows != nil {
			log.Print("Ошибка парсинга ответа БД: ", " (", err, ")")
			w.WriteHeader(http.StatusInternalServerError)
			// Комментарий: это очень некорректный ответ, но пока так.
			w.Write([]byte(err.Error()))
			return errRows
		}
	}

	// Комментарий: это очень оптимистичный и упрощенный вариант
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ok"))

	// Комментарий: Я предпочитаю явно указывать на отсутствие ошибки, в случае правильного прохождения кода
	return nil
}

// P.S. Есть классная статья, в которой описано почему не стоит использовать стандартный http пакет
// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779


// 2. Напишите короткий пример илюстрирующий проблему динамических масивов в GoLang.




// 3. Найдите ошибку в тексте, все используемы объекты загружены:

// 3.1. Разбор

type Service struct {

}

func (s *Service) Parents(category *model.Category) ([]*model.Category, error) {
	var categories []*model.Category

	cur := category

	// Комментарий: сам ParentID может быть `0` (такой индекс может быть), поэтому я бы делал ParentID, как указатель,
	// чтобы пустое значение было `nil`, а не `0`
	for cur.ParentID > 0 {
		cur, err := cur.Parent()
		if err != nil {
			return nil, err
		}
		// Ошибка: как cur, может вернуться nil, тогда нужно сделать break
		categories = append(categories, cur)
	}

	// Ошибка: нужно проверять заранее длину массива, иначе `j` будет -1, что сделает `i` всегда больше `j`
	for i, j := 0, len(categories)-1; i < j; i, j = i+1, j-1 {
		categories[i], categories[j] = categories[j], categories[i]
	}

	return categories, nil
}

// 3.2. Коррекция

type Category struct {
	ParentID *int
}

func (cat *Category) Parent() (*Category, error) {
	// Комментарий: это исключительно для примера
	return nil, nil
}

func (s *Service) ParentsCorrect(category *Category) ([]*Category, error) {
	var categories []*Category

	cur := category

	for cur.ParentID != nil {
		cur, err := cur.Parent()
		if err != nil {
			return nil, err
		}
		if cur == nil {
			break
		}
		categories = append(categories, cur)
	}

	if len(categories) > 0 {
		for i, j := 0, len(categories)-1; i < j; i, j = i+1, j-1 {
			categories[i], categories[j] = categories[j], categories[i]
		}
	}

	return categories, nil
}

func main() {

}
