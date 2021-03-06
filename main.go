package main

import (
	"net/http"
	"log"
	"encoding/json"
	"fmt"
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
// Если сюда уже передается *http.Request в неразобранном виде, могу предположить, что и http.ResponseWriter требуется передать.
func object_input_insert(r *http.Request) (error, string) {
	// Ошибка: указывается переменная, которой присваивается значение,
	// но оно нигде не использутеся. Требуется убрать.
	var flag string
	// Ошибка: ранее переменная не была определена, поэтому требуется использовать символ `:=`
	// Ошибка: сюда не передается ничего из запроса, это не совсем корректно
	// Вопрос: а почему вообще вставляется `id` в базу данных? Можно делать кастомные id,
	// чтобы исключить возможность иттерации по базе из вне или скрыть количество записей.
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
		// Комментарий: странно, что результат выводится в ошибке, лучше изменить эти строки.
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


// Комментарий:
func object_input_insert_correct(w http.ResponseWriter,r *http.Request) error {
	// Комментарий: поскольку здесь производится `insert`, я предполагаю, что это POST запрос, у которого есть тело.
	// Если это не POST, то роутер не пропустит запрос до данного уровня, поэтому здесь не требуется валидация.
	// Комментарий: также, я предположу, что используется JSON. Чаще всего для декодинга я использую
	// библиотеку https://github.com/mailru/easyjson
	dc := json.NewDecoder(r.Body)
	var reqB TestJSON
	if err := dc.Decode(&reqB); err != nil {
		log.Print("Ошибка парсинга запроса (", err, ")!")
		return err
	}
	defer r.Body.Close()

	// Комментарий: чаще всего, в каждой библиотеке есть уже встроенная функция создания запроса с аргументами,
	// которая сделает грамотный escape и все нужные обработки, поэтому писать свою реализацию считаю неправильным
	// (долгим, с возможностью упустить что-то или допустить ошибки. Готовые решения в этом плане проверенны тысячи раз)
	resp, err := db.Query("insert into test (text) values ($1)", reqB.Text)

	if err != nil {
		log.Print("Ошибка взаимодействия с БД (", err, ")!")
		// Комментарий: статус и текст ошибки будут зависеть от типа самой ошибки, а значит зависит от библиотеки.
		// Если проблема с коннектом к ДБ, то это 500 и проблема сервера + вызов дополнительных функций срочного
		// оповещения админов, если неккоректные данные, то это уже 422 или нечто похожее
		w.WriteHeader(http.StatusBadRequest)
		// Комментарий: это очень некорректный ответ, но пока так. Обычно я делаю единную структуру по которой возвращаю ошибку.
		w.Write([]byte(err.Error()))
		// Комментарий: Я предпочитаю явно возвращать ошибку, в месте, где она должна оставновить код
		return err
	} else {
		_, errRows := resp.RowsAffected()
		if errRows != nil {
			log.Print("Ошибка парсинга ответа БД: ", " (", err, ")")
			w.WriteHeader(http.StatusInternalServerError)
			// Комментарий: это очень некорректный ответ, но пока так. Обычно я делаю единную структуру по которой возвращаю ошибку.
			w.Write([]byte(err.Error()))
			return errRows
		}
	}

	// Комментарий: это очень оптимистичный и упрощенный вариант ответа.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ok"))

	// Комментарий: Я предпочитаю явно указывать на отсутствие ошибки, в случае правильного прохождения кода
	return nil
}

// P.S. Есть классная статья, в которой описано почему не стоит использовать стандартный http пакет
// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779




// 2. Напишите короткий пример илюстрирующий проблему динамических масивов в GoLang.

// 2.1. Разбор

// Мне не совсем понятна формулировка "проблема" динамических массивов. Любая вещь в языке программирования,
// которая не заявлена разработчиками, как ошибка реализации, является "особенностью" языка. Но для конкретного
// разработчика эта "особенность" может быть "проблемой" (из-за опыта в других языках, невозможности запомнить или понять и т.п.).
// Итогово, "проблема" понятие относительное, поэтому я могу описать особенности динамических массивов.
// + Я искал возможную слишком тонкую особенность, которую можно было бы назвать "проблемой", но ничего не нашел и сам не встречал

// 2.2. Решение

// Возможно, самой необычной вещью явлется это:
// При каждом добавлении с append, если длина изначального массива недостаточна (емкость среза),
// то создается новый массив с увеличенной в 2 раза величиной. Поэтому получается, что пока длинна массива достаточна,
// изменения происходят в самом массиве. Также append возвращает новый срез, поэтому его нужно куда-то присваивать
// (чаще всего в переменную целевого среза).

// И да, это поведение не контролируемое и если требуется реализации другого изменения размерности массива или уверенности
// в будет ли на выходе тот же массив или создасться новый, потребуется написать свою функцию реализации добавления в массив

func testSliceExpand() {
	arr := [3]int{ 1,2,3 }
	slCap := arr[0:2]
	fmt.Println(len(arr))
	fmt.Println(arr)
	fmt.Println(slCap)
	slCap[1] = 6
	slCap = append(slCap, 4, 5) // Создался новый массив и изменения теперь будут происходить именно в новом массиве,
	// причем его длинна будет 8, а не 6
	fmt.Println(len(arr))
	fmt.Println(arr)
	fmt.Println(slCap)
	slCap[1] = 8
	fmt.Println(len(arr))
	fmt.Println(arr)
	fmt.Println(slCap)
}

// . Еще одна важная вещь: нужно помнить что функция коппирования скопирует ссылки на те же элементы и все внутренние ссылки тоже
// в стандарте нет функции deepCopy (вот это может быть проблемой)

// Я выделю какие-то основные и тонкие моменты:

// . Массивы (array) как таковые в Golang всегда имеют статическую длину и сам массив не может быть расширен напрямую
// Пример создания массива

func arrCreate() {
	var arr [22]int // Пустой массив длинной в 22 элемента, каждый из которых встает в нулевое значение типа (в данном случае в `0`)
	// Для записи в массив используется указание индекса и значение
	arr[4] = 5
	// Для иттерации используется for или range
	for i := 0; i < len(arr); i++ {
		fmt.Println(i, arr[i])
	}
	for i, val := range arr {
		fmt.Println(i, val)
	}
	testArr := [3]int{ 2,5,6 } // Массив с установленными значениями
	fmt.Println(testArr)
}

func theSameArray() {
	scores := make([]int, 100)
	m := scores[3:50]
	m[20] = 1
	m2 := m[10:30]
	m2[19] = 2
	fmt.Println(scores) // все изменения ("1" и "2") будут отражены в данном срезе
	fmt.Println(m)
	fmt.Println(m2)
}

// . Динамической длины может быть срез (slice), представляющий собой грубо говоря "ссылку на отрезок массива", то есть
// любой отрезок всегда под собой имеет массив, изменение в этом отрезке, в открезках от первоначального отрезка,
// отрезках созданных на основании того же самого массива, всегда будут менять именно первоначальный массив.
// У slice есть две характеристики: длинна самого slice и емкость (длинна массива, на который ссылается slice)

func sliceCreate() {
	var slEmpty []int // Пустой slice с длинной в 0 и емкостью в 0
	slFull := []string{"hello", "world"} // Заполненный срез с длинной и емкостью в 2 элемента
	slLengthAndCap := make([]bool, 10) // Пустой срез длинной и емкостью в 10 элементов
	// make используется, поскольку для создания данного среза потребуется заранее
	// инициализировать массив внутри самого Go
	slCap := make([]int, 0, 10) // Пустой срез длинной в 0 и емкостью в 10 элементов
	var arr [10]int
	slFromArr := arr[3:5] // Сред по массиву

	// Второй, третий и пятый случай можно использовать с индексами.
	// А первый и четвертый только с оператором append, поскольку они имеют нулевую длину.
	fmt.Println(slEmpty)
	fmt.Println(slFull)
	fmt.Println(slLengthAndCap)
	fmt.Println(slCap)
	fmt.Println(slFromArr)
}




// 3. Найдите ошибку в тексте, все используемы объекты загружены:

// 3.1. Разбор

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

type Service struct {

}

// Комментарий: прописал здесь, чтобы не выделять в отдельный пакет `model` для упрощения.
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
