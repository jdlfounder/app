# app
operator-sdk test


创建 APP:
	创建工程目录和设置go下载包的代理:
	export GOPROXY=https://goproxy.cn,direct GOPATH=/www/data/go
	mkdir -p $GOPATH/src/github.com/jdlfounder/app && $GOPATH/src/github.com/jdlfounder/app
	
	创建脚手架工程:
		operator-sdk init --domain example.com --repo=github.com/jdlfounder/app --license apache2 --owner "jdlfounder" --component-config true
      
  创建 API:
		operator-sdk create api --group app --version v1 --kind App --resource=true --controller=true
    
  修建控制器的镜像
			make docker-build  IMG=gouqs/app:v1.1.0
		1>推送控制器镜像到hub.docker.com：
			登录hub.docker.com：
				docker login -u gouqs -p j*@1*5
			推送:
				make docker-push  IMG=gouqs/app:v1.1.0
			说明：
				如果不同送该镜像到gouqs的账户下，下面的部署不会成功，因为下面的部署方式要求去hub.docker.com的gouqs账户下拉取镜像:app:v1.1.0
		2>部署
			make deploy  IMG=gouqs/app:v1.1.0
		3>创建自定义资源:
			kubectl apply -f config/samples/app_v1_app.yaml 
		4>删除部署:
			删除 CRD 资源
				kubectl apply -f config/samples/app_v1_app.yaml 
				
			删除 CRD 定义
				make uninstall
			删除 controller
				make undeploy

			其实一条语句就搞定:
				make  undeploy  IMG=gouqs/app:v1.1.0
        
       错误解决：
	W0412 09:49:54.991967       1 reflector.go:324] pkg/mod/k8s.io/client-go@v0.23.0/tools/cache/reflector.go:167: failed to list *v1.Deployment: deployments.apps is forbidden: User 		"system:serviceaccount:app-system:app-controller-manager" cannot list resource "deployments" in API group "apps" at the cluster scope
	E0412 09:49:54.992011       1 reflector.go:138] pkg/mod/k8s.io/client-go@v0.23.0/tools/cache/reflector.go:167: Failed to watch *v1.Deployment: failed to list *v1.Deployment: 		deployments.apps is forbidden: User "system:serviceaccount:app-system:app-controller-manager" cannot list resource "deployments" in API group "apps" at the cluster scope
	权限问题，自动定义的集群角色app-manager-role只服用了服务帐户app-controller-manager在命名空间app-system下访问app.example.com api下的资源，
	而这里需要在default命名空间创建deployments和pod等资源。根本原因在于能否的api 组只有app.example.com。
	解决办法：
		kubectl  edit clusterrole app-manager-role -o yaml
		添加下面的规则:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - '*'

	 修改后pod起来了：

		kubectl  get pods 
		NAME                          READY   STATUS    RESTARTS   AGE
		app-sample-5d665d6b96-45pdg   1/1     Running   0          4m59s
		app-sample-5d665d6b96-4697h   1/1     Running   0          4m59s


		注意下面的规则实际上赋予了app-system命名空间下的服务帐户app-controller-manager非常高的权限了，生产中避免授予这么高的权限。
