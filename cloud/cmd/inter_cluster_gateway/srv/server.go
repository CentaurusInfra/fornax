package srv

import (
	"context"
	"errors"
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
	"github.com/kubeedge/kubeedge/common/constants"
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
	gatewayConfig, err := s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Get(context.TODO(), constants.ClusterGatewayConfigMap, metav1.GetOptions{})
	if err != nil {
		return &proto.Response{ReturnCode: proto.Response_Error}, err
	}
	updatedVpcGateways := make([]string, 0)
	if vpcGateways, ok := gatewayConfig.Data[constants.ClusterGatewayConfigMapVpcGateways]; ok && len(vpcGateways) > 0 {
		vpcGatewayArr := strings.Split(vpcGateways, ",")
		for _, vpcGateway := range vpcGatewayArr {
			vpcGatewayPair := strings.Split(vpcGateway, "=")
			if vpcGatewayPair[0] == request.GetName() && vpcGatewayPair[1] == request.GetGatewayHostIP() {
				return &proto.Response{ReturnCode: proto.Response_OK, Message: "The vpc gateway has already been added"}, nil
			}
			updatedVpcGateways = append(updatedVpcGateways, vpcGateway)
		}
	}
	updatedVpcGateways = append(updatedVpcGateways, fmt.Sprintf("%s=%s", request.GetName(), request.GetGatewayHostIP()))
	gatewayConfig.Data[constants.ClusterGatewayConfigMapVpcGateways] = strings.Join(updatedVpcGateways, ",")

	_, err = s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Update(context.TODO(), gatewayConfig, metav1.UpdateOptions{})
	if err != nil {
		return &proto.Response{ReturnCode: proto.Response_Error}, err
	}
	return &proto.Response{ReturnCode: proto.Response_OK}, nil
}

func (s *server) DeleteVpcGateway(ctx context.Context, request *proto.DeleteVpcGatewayRequest) (*proto.Response, error) {
	gatewayConfig, err := s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Get(context.TODO(), constants.ClusterGatewayConfigMap, metav1.GetOptions{})
	if err != nil {
		return &proto.Response{ReturnCode: proto.Response_Error}, err
	}
	if vpcGateways, ok := gatewayConfig.Data[constants.ClusterGatewayConfigMapVpcGateways]; ok {
		updatedVpcGateways := make([]string, 0)
		vpcGatewayArr := strings.Split(vpcGateways, ",")
		for _, vpcGateway := range vpcGatewayArr {
			vpcGatewayPair := strings.Split(vpcGateway, "=")
			if vpcGatewayPair[0] != request.GetName() || vpcGatewayPair[1] != request.GetGatewayHostIP() {
				updatedVpcGateways = append(updatedVpcGateways, vpcGateway)
			}
		}

		if len(vpcGatewayArr) != len(updatedVpcGateways) {
			gatewayConfig.Data[constants.ClusterGatewayConfigMapVpcGateways] = strings.Join(updatedVpcGateways, ",")
			_, err = s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Update(context.TODO(), gatewayConfig, metav1.UpdateOptions{})
			if err != nil {
				return &proto.Response{ReturnCode: proto.Response_Error}, err
			}
		}
	}
	return &proto.Response{ReturnCode: proto.Response_OK}, nil
}

func (s *server) CreateSubnet(ctx context.Context, request *proto.CreateSubnetRequest) (*proto.Response, error) {
	if subnet, err := s.subnetClientset.MizarV1().Subnets(request.GetNamespace()).Get(context.TODO(), request.GetName(), metav1.GetOptions{}); err == nil && subnet != nil {
		return &proto.Response{ReturnCode: proto.Response_Error}, errors.New("duplicated subnet")
	}

	if gatewayConfig, err := s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Get(context.TODO(), constants.ClusterGatewayConfigMap, metav1.GetOptions{}); err == nil {
		updatedSubGateways := make([]string, 0)
		if subGateways, ok := gatewayConfig.Data[constants.ClusterGatewayConfigMapSubGateways]; ok && len(subGateways) > 0 {
			subGatewayArr := strings.Split(subGateways, ",")
			for _, subGateway := range subGatewayArr {
				subatewayPair := strings.Split(subGateway, "=")
				if subatewayPair[0] != request.GetName() {
					updatedSubGateways = append(updatedSubGateways, subGateway)
				}
			}
		}
		updatedSubGateways = append(updatedSubGateways, fmt.Sprintf("%s=%s", request.GetName(), request.GetRemoteGateway()))
		gatewayConfig.Data[constants.ClusterGatewayConfigMapSubGateways] = strings.Join(updatedSubGateways, ",")
		_, err = s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Update(context.TODO(), gatewayConfig, metav1.UpdateOptions{})
		if err != nil {
			return &proto.Response{ReturnCode: proto.Response_Error}, err
		}
	}
	subnet := &subnetv1.Subnet{}
	subnet.Name = request.GetName()
	subnet.Namespace = request.GetNamespace()
	subnet.Spec.IP = request.GetIP()
	subnet.Spec.Prefix = request.GetPrefix()
	subnet.Spec.Vni = request.GetVni()
	subnet.Spec.Vpc = request.GetVpc()
	subnet.Spec.Status = "Init"
	subnet.Spec.Bouncers = int(request.GetBouncers())
	subnet.Spec.Virtual = true
	_, err := s.subnetClientset.MizarV1().Subnets(subnet.Namespace).Create(context.TODO(), subnet, metav1.CreateOptions{})
	if err != nil {
		return &proto.Response{ReturnCode: proto.Response_Error}, err
	}
	return &proto.Response{ReturnCode: proto.Response_OK}, nil
}

func (s *server) DeleteSubnet(ctx context.Context, request *proto.DeleteSubnetRequest) (*proto.Response, error) {
	if gatewayConfig, err := s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Get(context.TODO(), constants.ClusterGatewayConfigMap, metav1.GetOptions{}); err == nil {
		updatedSubGateways := make([]string, 0)
		if subGateways, ok := gatewayConfig.Data[constants.ClusterGatewayConfigMapSubGateways]; ok && len(subGateways) > 0 {
			subGatewayArr := strings.Split(subGateways, ",")
			for _, subGateway := range subGatewayArr {
				subatewayPair := strings.Split(subGateway, "=")
				if subatewayPair[0] != request.GetName() {
					updatedSubGateways = append(updatedSubGateways, subGateway)
				}
			}
		}
		gatewayConfig.Data[constants.ClusterGatewayConfigMapSubGateways] = strings.Join(updatedSubGateways, ",")
		_, err = s.clientset.CoreV1().ConfigMaps(request.GetNamespace()).Update(context.TODO(), gatewayConfig, metav1.UpdateOptions{})
		if err != nil {
			return &proto.Response{ReturnCode: proto.Response_Error}, err
		}
	}
	err := s.subnetClientset.MizarV1().Subnets(request.GetNamespace()).Delete(context.TODO(), request.GetName(), metav1.DeleteOptions{})
	if err != nil {
		return &proto.Response{ReturnCode: proto.Response_Error}, err
	}
	return &proto.Response{ReturnCode: proto.Response_OK}, nil
}
