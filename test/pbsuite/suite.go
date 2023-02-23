package pbsuite

import (
	"net"
	"strconv"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Suite - реализация suite.Suite для тестирования gRPC
type Suite struct {
	suite.Suite
	HelloClient  HelloServiceClient
	AnswerClient AnswerServiceClient
	Server       *grpc.Server
	listener     net.Listener
	conn         *grpc.ClientConn
}

// TearDownSuite - метод интерфейса suite.TearDownAllSuite.
// Вызывается после завершения всех тестов.
func (suite *Suite) TearDownSuite() {
	// Останавливаем сервер (listener закрывается внутри server.Stop)
	if suite.Server != nil {
		suite.Server.Stop()
	}
	// Закрываем клиентское соединение
	if suite.conn != nil {
		_ = suite.conn.Close()
	}
}

// StartServer - запускает gRPC сервер и подключает к нему клиентов
func (suite *Suite) StartServer(server *grpc.Server) {
	// Останавливаем предыдущий сервер (listener закрывается внутри server.Stop)
	if suite.Server != nil {
		suite.Server.Stop()
	}
	// Закрываем предыдущее клиентское соединение
	if suite.conn != nil {
		_ = suite.conn.Close()
	}

	// Открываем листенер на случайном свободном порту
	var err error
	suite.listener, err = net.Listen("tcp", ":0")
	suite.Require().NoError(err)
	port := strconv.Itoa(suite.listener.Addr().(*net.TCPAddr).Port)

	// Создаем клиентское соединение к этому порту
	suite.conn, err = grpc.Dial(":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	suite.Require().NoError(err)

	// Запускаем сервер
	suite.Server = server
	go func() {
		suite.Require().NoError(suite.Server.Serve(suite.listener))
	}()

	// Регистрируем сервисы
	RegisterHelloServiceServer(suite.Server, &HelloService{})
	RegisterAnswerServiceServer(suite.Server, &AnswerService{})

	// Подключаем клиентов
	suite.HelloClient = NewHelloServiceClient(suite.conn)
	suite.AnswerClient = NewAnswerServiceClient(suite.conn)
}
