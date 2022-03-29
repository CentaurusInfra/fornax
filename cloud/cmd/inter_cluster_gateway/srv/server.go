package srv

import (
	"context"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	subnetclientset "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/subnet/client/clientset/versioned"
	subnetv1 "github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/pkg/apis/subnet/v1"
	"github.com/kubeedge/kubeedge/cloud/cmd/inter_cluster_gateway/srv/proto"
)

type server struct {
	kubeconfig      *rest.Config
	subnetClientset *subnetclientset.Clientset
	clientset       *kubernetes.Clientset
}

func RunGrpcServer(conf *rest.Config, port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	subnetClientset, err := subnetclientset.NewForConfig(conf)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(conf)
	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer()
	proto.RegisterMizarServiceServer(srv, &server{kubeconfig: conf, subnetClientset: subnetClientset, clientset: clientset})

	if e := srv.Serve(listener); e != nil {
		panic(e)
	}
}

func (s *server) CreateVpcGateway(ctx context.Context, request *proto.CreateVpcGatewayRequest) (*proto.Response, error) {
	gatewayConfig, err := s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Get(context.TODO(), "cluster-gateway-config", metav1.GetOptions{})
	if err != nil {
		return &proto.Response{ReturnCode: proto.Response_Error}, err
	}
	if gateways, ok := gatewayConfig.Data[request.GetName()]; ok {
		ipArr := strings.Split(gateways, ",")
		for _, ip := range ipArr {
			if ip == request.GetGatewayHostIP() {
				return &proto.Response{ReturnCode: proto.Response_OK, Message: "The vpc gateway has already been added"}, nil
			}
		}
		gatewayConfig.Data[request.GetName()] = fmt.Sprintf("%s,%s", gateways, request.GetGatewayHostIP())
	} else {
		gatewayConfig.Data[request.GetName()] = request.GetGatewayHostIP()
	}
	_, err = s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Update(context.TODO(), gatewayConfig, metav1.UpdateOptions{})
	if err != nil {
		return &proto.Response{ReturnCode: proto.Response_Error}, err
	}
	return &proto.Response{ReturnCode: proto.Response_OK}, nil
}

func (s *server) DeleteVpcGateway(ctx context.Context, request *proto.DeleteVpcGatewayRequest) (*proto.Response, error) {
	gatewayConfig, err := s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Get(context.TODO(), "cluster-gateway-config", metav1.GetOptions{})
	if err != nil {
		return &proto.Response{ReturnCode: proto.Response_Error}, err
	}
	if gateways, ok := gatewayConfig.Data[request.GetGatewayName()]; ok {
		gatewayArray := strings.Split(gateways, ",")
		updatedGatewayArray := make([]string, 0)
		for _, gateway := range gatewayArray {
			if gateway != request.GetGatewayHostIP() {
				updatedGatewayArray = append(updatedGatewayArray, gateway)
			}
		}
		if len(gatewayArray) != len(updatedGatewayArray) {
			gatewayConfig.Data[request.GetGatewayName()] = strings.Join(updatedGatewayArray, "-")
			_, err = s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Update(context.TODO(), gatewayConfig, metav1.UpdateOptions{})
			if err != nil {
				return &proto.Response{ReturnCode: proto.Response_Error}, err
			}
		}
	}
	return &proto.Response{ReturnCode: proto.Response_OK}, nil
}

func (s *server) CreateSubnet(ctx context.Context, request *proto.CreateSubnetRequest) (*proto.Response, error) {
	subnet := &subnetv1.Subnet{}
	subnet.Name = request.GetName()
	subnet.Namespace = request.GetNamespace()
	subnet.Spec.IP = request.GetIP()
	subnet.Spec.Prefix = request.GetPrefix()
	subnet.Spec.Vni = request.GetVni()
	subnet.Spec.Vpc = request.GetVpc()
	subnet.Spec.Status = request.GetStatus()
	subnet.Spec.Bouncers = 1
	subnet.Spec.Virtual = true
	_, err := s.subnetClientset.MizarV1().Subnets(subnet.Namespace).Create(context.TODO(), subnet, metav1.CreateOptions{})
	if err != nil {
		return &proto.Response{ReturnCode: proto.Response_Error}, err
	}
	return &proto.Response{ReturnCode: proto.Response_OK}, nil
}

func (s *server) DeleteSubnet(ctx context.Context, request *proto.DeleteSubnetRequest) (*proto.Response, error) {
	err := s.subnetClientset.MizarV1().Subnets(request.GetNamespace()).Delete(context.TODO(), request.GetName(), metav1.DeleteOptions{})
	if err != nil {
		return &proto.Response{ReturnCode: proto.Response_Error}, err
	}
	return &proto.Response{ReturnCode: proto.Response_OK}, nil
}
