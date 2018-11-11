## OpenFaaS Community

Welcome to the OpenFaaS community page where you can find:

* [Industry awards, notable mentions and books](#industry-awards-notable-mentions-and-books)
* [Function Providers/Back-ends](#openfaas-providers)

* Events
    * [2018](#events-in-2018)
    * [2017](#events-in-2017)

* Blog posts and write-ups
    * [2018](#blog-posts-and-write-ups-2018)
    * [2017](#blog-posts-and-write-ups-2017)
* [Repositories - Projects, Samples, and Use Cases](#repositories)

There are several key GitHub repositories for the project which are available under the [openfaas](https://github.com/openfaas) organisation. Incubator and experimental projects are under the [openfaas-incubator](https://github.com/openfaas-incubator) organisation.

#### How to submit a link to this page

It would be great to hear from you especially if you have any of the above and want to share it with the rest of the community. Pull Request or submit a Github Issue.

> Note: You will need to sign-off the commit too using `git commit -s` or add `Signed-off-by: {your full name} <{your email address}>` to the commit message when using the UI.

### Industry awards, notable mentions and books

[Back to top](#openfaas-community)

| Award/Mention                                          | Author(s)       | Source     | Date        |
|---------------------------------------------------------------------|--------------|----------|-------------|
| 🏆🏆 [2018 Bossie Award Best OSS Cloud Computing](https://www.infoworld.com/article/3306455/cloud-computing/the-best-open-source-software-for-cloud-computing.html#slide13) | Jonathan Freeman et al. | InfoWorld.com | 26-Sep-2018 |
| [KubeWeekly #145](https://mailchi.mp/cncf/kubeweekly-145?e=5c7c372824) | [KubeWeekly](https://twitter.com/kubeweekly?lang=en) | mailchi.mp | 15-Aug-2018 |
| [Book: Docker for Serverless Applications with OpenFaaS - featuring OpenFaaS](https://www.packtpub.com/mapt/book/virtualization_and_cloud/9781788835268) | Chanwit Kaewkasi | packtpub.com | 01-Apr-2018 |
| [The Cube interview at Cisco DevNet Create](https://www.youtube.com/watch?v=J8UYZ1GXNTQ) | The Cube | Youtube.com | 11-Apr-2018
| [VMware hires OpenFaaS Founder](https://thenewstack.io/vmware-hires-openfaas-founder-dell-drops-code-initiative-2/) | The New Stack | thenewstack.io | 21-Feb-2018 |
| [Book: Kubernetes for Serverless Applications - featuring OpenFaaS](https://www.packtpub.com/mapt/book/networking_and_servers/9781788620376) |  Russ McKendrick | packtpub.com | 01-Jan-2018
| [OpenFaaS added to CNCF Cloud Native Landscape - Serverless/Event-based](https://github.com/cncf/landscape/blob/master/landscape/CloudNativeLandscape_latest.jpg) | [Cloud Native Computing Foundation (CNCF)](https://www.cncf.io/) | Github | 23-Oct-2017 |
| [OpenFaaS: Package any Binary or Code as a Serverless Function](https://thenewstack.io/openfaas-put-serverless-function-container/) | Joab Jackson | The New Stack | 9-Oct-2017 |
| [Functions-as-a-Service Landscape](https://medium.com/memory-leak/this-year-gartner-added-serverless-to-its-hype-cycle-of-emerging-technologies-reflecting-the-5dfe43d818f0) | Astasia Myers | Redpoint Ventures | 4-Oct-2017 |
| 🏆 [Bossie Award for Cloud Computing 2017](https://www.infoworld.com/article/3227920/cloud-computing/bossie-awards-2017-the-best-cloud-computing-software.html#slide7) |  Martin Heller, Andrew C. Oliver and Serdar Yegulalp| InfoWorld (from IDG) |  27-Sep-2017 |
| [Open source project uses Docker for serverless computing](http://www.infoworld.com/article/3184757/open-source-tools/open-source-project-uses-docker-for-serverless-computing.html#tk.twt_ifw)| Serdar Yegulalp  | InfoWorld (from IDG)  | 27-Mar-2017 |

### OpenFaaS providers

[Back to top](#openfaas-community)

Official providers developed and supported by the OpenFaaS project

| Project name and description                                         | Author     | Site      | Status      |
|----------------------------------------------------------------------|------------|-----------|-------------|
| **faas-netes**- Kubernetes provider | OpenFaaS | [github.com](https://github.com/openfaas/faas-netes) | Supported |
| **faas-swarm** - Docker Swarm provider | OpenFaaS | [github.com](https://github.com/openfaas/faas-swarm) | Supported |
| **openfaas-operator** - Kubernetes Operator provider | OpenFaaS | [github.com](https://github.com/openfaas-incubator/openfaas-operator) | Incubation |

Community providers actively being developed and/or supported by a third-party

| Project name and description                                         | Author     | Site      | Status      |
|----------------------------------------------------------------------|------------|-----------|-------------|
| **faas-fargate** - AWS Fargate provider | Edward Wilde | [github.com](https://github.com/ewilde/faas-fargate) | Incubation |
| **faas-nomad** - Nomad provider | Nic Jackson (Hashicorp) | [github.com](https://github.com/hashicorp/faas-nomad) | Incubation  |
| **faas-rancher** - Rancher/Cattle provider | Ken Fukuyama | [github.com](https://github.com/kenfdev/faas-rancher) | Inception |
| **faas-dcos** - DCOS provider | Alberto Quario | [github.com](https://github.com/realbot/faas-dcos) | Inception  |
| **faas-hyper** - Hyper.sh provider | Hyper | [github.com](https://github.com/hyperhq/faas-hyper) | Inception |
| **faas-guardian** - Guardian provider | Nigel Wright | [github.com](https://github.com/nwright-nz/openfaas-guardian-backend) | Inception |
| **faas-ecs** | Xicheng Chang (Huawei) | [github.com](https://github.com/stack360/faas-ecs) | Inception |

Key:

(Status)

* Incubation - actively being developed, but not stable
* Inception - more work needed. May work in happy path scenario, but unlikely to be suitable for production

### 2018

#### Events in 2018
[Back to top](#openfaas-community)

| Event name and description                                          | Speaker      | Location | Date        |
|---------------------------------------------------------------------|--------------|----------|-------------|
| [KubeCon + CloudNativeCon: Digital Transformation of Vision Banco Paraguay with Serverless Functions](https://sched.co/GraO) | Alex Ellis & Patricio Diaz | Seattle, WA | 13-Dec-2018 |
| [The FaaS & The Furious](https://www.meetup.com/CloudAustin/events/gxwbkpyxpbbc/) | Burton Rheutan | Austin, TX | 27-Nov-2018 |
| [GOTO Copenhagen: Serverless Beyond the Hype](https://gotocph.com/2018/sessions/592) | Alex Ellis | Copenhagen, DK | 19-Nov-2018 |
| [OpenFaas on Kubernetes](https://www.meetup.com/kubernetes-openshift-India-Meetup/events/255794614) | Vivek Singh| Bangalore, IN | 17-Nov-2018 |
| [ Serverless Computing London: FaaS and Furious – 0 to Serverless in 60 Seconds, Anywhere](https://serverlesscomputing.london/sessions/faas-furious-0-serverless-60-seconds-anywhere/) | Alex Ellis | London, UK | 12-Nov-2018 |
| [Open Source Summit Europe: Introducing OpenFaaS Cloud, a Developer-Friendly CI/CD Pipeline for Serverless](https://osseu18.sched.com/event/FxXx/introducing-openfaas-cloud-a-developer-friendly-cicd-pipeline-for-serverless-alex-ellis-openfaas-project-vmware) | Alex Ellis | Edinburgh, Scotland | 22-Oct-2018 |
| Exploring the serverless framework with OpenFaaS @ DevFest Nantes | [Emmanuel Lebeaupin](https://twitter.com/elebeaup) | Nantes, France | 19-Oct-2018 |
| [Serverlessconf Tokyo 2018 Contributor Day](https://serverless.connpass.com/event/103404/) | Ken Fukuyama | Tokyo, JP | 13-Oct-2018 |
| [DevOpenSpace - Mühelos Serverless mit GitOps](https://github.com/fpommerening/DevOpenSpace2018) | Frank Pommerening | Leipzig, Germany | 12-Oct-2018 |
| [ScotlandPHP - Starting your Serverless Journey with OpenFaaS](https://conference.scotlandphp.co.uk/) | John McCabe | Edinburgh, Scotland | 6-Oct-2018 |
| [VMUG Ireland - Serverless, OpenFaaS and GitOps on Kubernetes](https://community.vmug.com/events/event-description?CalendarEventKey=081889a6-3ea0-451b-b5df-e3985fbe8256&CommunityKey=8c80c6de-4f79-4717-b7d1-55a43ba44624) | John McCabe | Dublin, Ireland | 4-Oct-2018 |
| [Serverlessconf Tokyo 2018 OpenFaaS Workshop](http://tokyo.serverlessconf.io/workshops.html) | Ken Fukuyama | Tokyo, JP | 28-Sep-2018 |
| [Building Cloud-Native Apps - FaaS: The Swiss Army Knife of Cloud Native](https://www.meetup.com/Cloud_Native_Meetup/events/254232936/) | James McAfee | Sydney, Australia | 12-Sep-2018 |
| [ John Callaway: .NET Core on a Raspberry Pi Cluster with Docker and OpenFaaS](https://secure.meetup.com/register/?ctx=ref) | John Callaway | Orlando, FL, USA |16-Aug-2018 |
| [Cloud Native Glasgow - Serverless on Kubernetes](https://www.meetup.com/CloudNativeGlasgow/events/251913059/) | John McCabe | Glasgow, UK | 5-July-2018 |
| [Opsview & Kubernetes (incl. OpenFaaS)](https://www.meetup.com/Birmingham-digital-development-Meetup/events/250680464/) | Opsview | Birmingham, UK | 28-June-2018 |
| [Meetup[4] - Pizza and OpenFaaS](https://www.meetup.com/Shirley-Software-Development-Meetup/events/251103825/) | [Steven Tsang](https://twitter.com/stetsang) | Shirley, UK | 27-June-2018 |
| [Open Source Sharing is Caring Meetup](https://www.meetup.com/VMwareBulgaria/events/251794762/) | Alex Ellis & Ivana Yovcheva | Sofia, Bulgaria | 25-June-2018 |
| [OpenFaaS Bangalore Launch Meetup](https://www.meetup.com/Bangalore-OpenFaaS-Meetup/events/251711771/) | Vivek Singh, Vivek Sridhar, Tarun Mangukiya | Bangalore, IN | 23-June-2018 |
| [Docker Seattle - Serverless Talks: Fn Project && OpenFaaS](https://www.meetup.com/Docker-Seattle/events/251686931/) | Eric Stoekl | Seattle, WA | 21-June-2018 |
| [OpenFaaS: Serverless Kubernetes Functions](https://www.meetup.com/meetup-group-hPTHYTAe/events/251435215/) | Burton Rheutan, Tunde Oladipupo | Zoom | 20-June-2018 |
| [GitOps, OpenFaaS, Istio and Helm with Weaveworks](https://www.meetup.com/Weave-User-Group-Bay-Area/events/251409954/) | Alex Ellis, Stefan Prodan | SF, USA | 12-June-2018 |
| [Serverless Panel @ Dockercon](https://dockercon18.smarteventscloud.com/connect/sessionDetail.ww?SESSION_ID=194869) | Alex Ellis & others | SF, USA | 14-June-2018 |
| [Container Innovation SIG @ Dockercon](https://dockercon18.smarteventscloud.com/connect/sessionDetail.ww?SESSION_ID=187204) | Alex Ellis & others | SF, USA | 13-June-2018 |
| [Serverless Beyond the Hype @ Docker London](https://www.meetup.com/Docker-London/events/249221771/) | Alex Ellis | London, UK | 25-May-2018 |
| [OpenFaaS @ GOTO Nights](https://www.meetup.com/GOTO-Nights-CPH/events/249895973/) | Alex Ellis | Copenhagen, DK | 1-May-2018 |
| [Kubernetes Seattle: OpenFaas & Migrating to Envoy Proxy in K8s](https://www.meetup.com/Seattle-Kubernetes-Meetup/events/250105287/) | Eric Stoekl | Seattle, WA | 23-May-2018 |
| [OpenFaaS @ Serverless CPH](https://serverlesscph.dk) | John Mccabe | Copenhagen, DK | 16-May-2018 |
| Serverless OpenFaaS and Python Workshop @ Agile Peterborough        | Alex Ellis, Richard Gee  | Peterborough, UK | 12-May-2018 |
| [Going Serverless with OpenFaaS, Kubernetes, and Python @ PyCon](https://us.pycon.org/2018/schedule/presentation/40/) | [Michael Herman](https://twitter.com/mikeherman) |  Cleveland, OH | 9-May-2018 |
| [OpenFaaS talk and workshop @ Cisco DevNet Create](https://devnetcreate.io/2018/pages/speaker/speaker.html#Alex-Ellis) | Alex Ellis | Mountain View, CA | 10/11-Apr-2018 |
| [OpenFaaS : a serverless framework on top of Docker and Kubernetes @ Devoxx France](https://www.devoxx.fr/) | [Laurent Grangeau](https://twitter.com/laurentgrangeau) | Paris, France | 18-Apr-2018 |
| [Deep Dive: Serverless Computing mit OpenFaaS - Magdeburger Developer Days](https://www.md-devdays.de/Act?id=34/) | Frank Pommerening | Magdeburg Germany | 10-April-2018 |
| [.NET Core on a Raspberry Pi Cluster with Docker and OpenFaaS](https://www.meetup.com/St-Pete-NET-Meetup/events/247299483/) | John Callaway | St Petersburg, FL | 03-Apr-2018 |
| [.NET Core on a Raspberry Pi Cluster with Docker and OpenFaaS](http://www.codepalousa.com/Sessions/1034) | John Callaway | Louisville, KY | 30-Mar-2018 |
| [OpenFaaS : a serverless framework on top of Docker and Kubernetes @ BreizhCamp](http://www.breizhcamp.org/) | [Laurent Grangeau](https://twitter.com/laurentgrangeau) | Paris, France | 29-Mar-2018 |
| [Running private OpenFaaS is not about avoiding lock-in](https://www.meetup.com/Serverless-Italy/events/248536634/) | Nic Jackson | Milan, IT | 20-Mar-2018 |
| [Serverless Computing mit Docker Swarm und OpenFaaS - Softwerkskammer Jena](https://www.meetup.com/de-DE/jenadevs/events/246045796/) | Frank Pommerening | Jena Germany | 20-Mar-2018 |
| [.NET Core on a Raspberry Pi Cluster with Docker and OpenFaaS](https://orlandocodecamp.com/sessions/Details/63) | John Callaway | Orlando, FL | 17-Mar-2018 |
| [Workshop: Serverless Computing mit OpenFaaS - Spartakiade](https://github.com/fpommerening/Spartakiade2018-OpenFaaS) | Frank Pommerening | Berlin Germany | 17-Mar-2018 |
| Serverless OpenFaaS and Python Workshop @ PyCon SK        | Alex Ellis  | Bratislava, SK | 11-Mar-2018 |
| Serverless with OpenFaaS Opening Keynote @ PyCon SK        | Alex Ellis  | Bratislava, SK | 09-Mar-2018 |
| [Serverless con AWS Lambda y OpenFaaS](https://t3chfest.uc3m.es/2018/programa/serverless-con-aws-lambda-openfaas/) | [Javier Revillas](https://twitter.com/javirevillas) | Madrid, Spain | 01-Mar-2018 |
| [OpenFaaS presentation + Demo on Kubernetes @ ClusterEurope](https://clustereurope.org/) | [Laurent Grangeau](https://twitter.com/laurentgrangeau) | Paris, France | 02-Feb-2018 |
| [OpenFaaS presentation + Demo on Kubernetes @ ApiDays Paris](http://www.apidays.io/events/paris-2017) | [Laurent Grangeau](https://twitter.com/laurentgrangeau) | Paris, France | 30-Jan-2018 |
| [Serverless Computing mit OpenFaaS - Developer Group Leipzig](https://www.meetup.com/de-DE/Developer-Group-Leipzig/events/245252994/) | Frank Pommerening | Leipzig Germany | 29-Jan-2018 |
| [OpenFaaS : a serverless framework on top of Docker and Kubernetes @ Snowcamp.io](https://snowcamp2018.sched.com/event/D2nX/openfaas-a-serverless-framework-on-top-of-docker-and-kubernetes?iframe=no&w=100%&sidebar=no&bg=no) | [Laurent Grangeau](https://twitter.com/laurentgrangeau) | Paris, France | 25-Jan-2018 |
| [Talk: Building a Raspberry Pi Kubernetes Cluster with OpenFaaS](https://ndc-london.com/talk/building-a-raspberry-pi-kubernetes-cluster-and-running-.net-core/) | Alex Ellis / Scott Hanselman | London | 18-Jan-2018 |

#### Blog posts and write-ups 2018
[Back to top](#openfaas-community)

| Blog/repo name and description                                          | Author       | Site     | Date        |
|---------------------------------------------------------------------|--------------|----------|-------------|
| [Using GraphQL in OpenFaaS functions](https://padiazg.github.io/graphql-on-opesfaas/) | Patricio Díaz | padiazg.github.io | 03-Nov-2018 |
| [🇬🇧-use-express-with-openfaas](https://k33g.gitlab.io/articles/2018-10-28-OPENFAAS.html) | Philippe Charrière | GitLab.com | 28-Oct-2018 |
| [First contact with OpenFaaS](https://k33g.gitlab.io/articles/2018-10-20-OPENFAAS.html) | Philippe Charrière | GitLab.com | 20-Oct-2018 |
| [ACI Information Group: OpenFaaS among wish-list items for AWS Lambda](https://searchaws.techtarget.com/opinion/Lambda-support-for-OpenFaaS-among-wish-list-items) | Chris Moyer | techtarget.com | 11-Oct-2018 |
| [Deploy a python app using OpenFaaS](https://saurabhlondhe.github.io/technology/linux/2018/10/10/python-app-with-openfaas.html) | Saurabh Londhe | saurabhlondhe.github.io | 10-Oct-2018 |
| [Serverless R functions with OpenFaaS](https://medium.com/@keyunie/serverless-r-functions-with-openfaas-1cd34905834d) | Keyu Nie | medium.com | 30-Sept-2018 |
| [Serverless Data Fetching stack with OpenFaaS & OpenFlow](https://blog.lapw.at/serverless-data-fetching-stack-openfaas-openflow/) | Quentin Lapointe | blog.lapw.at | 23-Sept-2018 |
| [How To Install and Secure OpenFaaS Using Docker Swarm on Ubuntu 16.04](https://www.digitalocean.com/community/tutorials/how-to-install-and-secure-openfaas-using-docker-swarm-on-ubuntu-16-04) | Marko Mudrinić | digitalocean.com | 19-Sept-2018 |
| [Vert.x + Graal = Java for Serverless = ❤️!](https://www.jetdrone.xyz/2018/09/07/Vertx-Serverless-5mb.html) | Paulo Lopes | jetdrone.xyz | 07-Sept-2018 |
| [OpenFass 介绍及源码分析 (OpenFaaS Deep-Dive / Code-Analysis) - Chinese](https://www.youtube.com/watch?v=bZtgrAVR9HQ&feature=youtu.be) | Zhu Zhenfeng | youtube.com | 07-Sept-2018 |
| [Introducing stateless microservices for OpenFaaS](https://www.openfaas.com/blog/stateless-microservices/) | Alex Ellis | openfaas.com | 05-Sept-2018 |
| [Serverless com OpenFaaS e PHP (Portuguese)](https://medium.com/@FernandoDebrand/serverless-com-openfaas-e-php-3dce499f8062) | Fernando Silva | medium.com | 04-Sept-2018 |
| [Deploy OpenFaaS and Kubernetes on DigitalOcean with Ansible](https://www.openfaas.com/blog/deploy-digitalocean-ansible/) | Richard Gee | openfaas.com | 27-Aug-2018 |
| [Cerner ShipIt Hackathon - ShipIt XII](https://engineering.cerner.com/blog/shipit-xii/) | Caitie Oder | cerner.com | 30-Aug-2918 |
| [5 tips and tricks for the OpenFaaS CLI](https://www.openfaas.com/blog/five-cli-tips/) | Alex Ellis | openfaas.com | 21-Aug-2018 |
| [Multi-stage Serverless on Kubernetes with OpenFaaS and GKE](https://www.openfaas.com/blog/gke-multi-stage/) | Stefan Prodan | openfaas.com | 14-Aug-2018 |
| [Managing state for Serverless Functions with Minio](https://blog.lapw.at/stateful-serverless-stack-openfaas-minio/) | Quentin Lapointe | lapw.at | 14-Aug-2018 |
| [Monitor Arista Cloud vision network devices with OpenFaaS](https://github.com/burnyd/cvp-serverless-openfaas) | Daniel Hertzberg | github.com | 13-Aug-2018 |
| [Serverless email sender using OpenFaaS and .NET Core](https://medium.com/@paulius.juozelskis/serverless-email-sender-using-openfaas-and-net-core-afdef152359) | Paulius Juozelskis | medium.com | 13-Aug-2018 |
| [VMware CodeHouse: A Great Insight into the Future of Software Engineering](https://blogs.vmware.com/opensource/2018/08/07/vmware-codehouse-empowers-women/) | Jonas Rosland | blogs.vmware.com | 7-Aug-2018 |
| [OpenFaaS on Cisco DevNet Sandbox](https://developer.cisco.com/docs/sandbox/#!cloud/all-cloud-sandboxes) | [Cisco DevNet Sandbox](https://developer.cisco.com/sandbox) | [Cisco DevNet](https://developer.cisco.com) | 8-Aug-2018 |
| [How to run Rust in OpenFaaS](https://booyaa.wtf/2018/run-rust-in-openfaas/) | Mark Sta Ana | booyaa.wtf | 4-Aug-2018 |
| [Scale to Zero and Back Again with OpenFaaS](https://www.openfaas.com/blog/zero-scale/) | Alex Ellis | openfaas.com | 25-July-2018 |
| [PHP and OpenFaaS](https://mfyu.co.uk/post/php-and-openfaas) | Matt Brunt | mfyu.co.uk | 21-July-2018 |
| [Deploying OpenFaaS from Portainer.io and Creating your First Function](https://www.linkedin.com/pulse/deploying-openfaas-from-portainerio-creating-your-first-steven-kang/) | Steven Kang | linkedin.com | 8-Jul-2018 |
| [Building a custom Dynamics 365 data interface with OpenFaaS](https://alexanderdevelopment.net/post/2018/07/05/building-a-custom-dynamics-365-data-interface-with-openfaas/) | Lucas Alexander | alexanderdevelopment.net | 5-Jul-2018 |
| [Introducing the OpenFaaS Operator for Serverless on Kubernetes](https://blog.alexellis.io/introducing-the-openfaas-operator/) | Alex Ellis | alexellis.io | 2-Jul-2018 |
| [Kubernetes: From Fear to Functions in 20 Minutes](https://medium.com/@rorpage/kubernetes-from-fear-to-functions-in-20-minutes-a021f7ef5843) | Rob Page | medium.com | 11-Jun-2018 |
| [Installing and securing OpenFaaS on an AKS cluster](https://alexanderdevelopment.net/post/2018/05/31/installing-and-securing-openfaas-on-an-aks/) | Lucas Alexander | alexanderdevelopment.net | 31-May-2018 |
| [How OpenFaaS came to rescue us!](https://medium.com/iconscout/how-openfaas-came-to-rescue-us-ec129518cd46) | Tarun Mangukiya | iconscout.com | 24-May-2018 |
| [Building a Serverless Microblog in Ruby with OpenFaaS – Part 1 ](https://keiran.scot/2018/05/18/faasfriday-serverless-microblog/) | Keiran Smith | keiran.scot | 18-May-2018 |
| [Using ML.NET in an OpenFaaS function](https://alexanderdevelopment.net/post/2018/05/17/using-ml-net-in-an-openfaas-function/) | Lucas Alexander | alexanderdevelopment.net | 17-May-2018 |
| [Demystifying Serverless & OpenFaas](https://collabnix.com/demystifying-openfaas-simplifying-serverless-computing/) | Ajeet Raina | collabnix.com | 27-Apr-2018
| [Writing OpenFaas Serverless Functions in Go](https://marcussmallman.io/2018/04/11/writing-openfaas-serverless-functions-in-go/) | Marcus Smallman | marcussmallman.io | 11-Apr-2018 |
| [Deploying OpenFaaS on Kubernetes - Azure AKS Part 2 (SSL)](https://medium.com/@ericstoekl/openfaas-on-azure-aks-with-ssl-ingress-openfaas-on-aks-part-2-2-e9f2db9db387) | Eric Stoekl | medium.com | 23-Mar-2018 |
| [VMware + OpenFaaS – One Month In](https://blogs.vmware.com/opensource/2018/03/20/vmware-openfaas-alex-ellis) | Alex Ellis | vmware.com | 21-Mar-2018 |
| [Securing Kubernetes for OpenFaaS and beyond](https://www.twistlock.com/2018/03/20/securing-kubernetes-openfaas-beyond) | Daniel Shapira | twistlock.com | 20-Mar-2018 |
| [What Is OpenFaaS and How Can It Drive Innovation?](https://www.contino.io/insights/what-is-openfaas-and-why-is-it-an-alternative-to-aws-lambda-an-interview-with-creator-alex-ellis) | Ben Tannahill | contino.com | 19-Mar-2018 |
| [Installing and securing OpenFaaS on a Google Cloud virtual machine](https://alexanderdevelopment.net/post/2018/02/25/installing-and-securing-openfaas-on-a-google-cloud-virtual-machine/) | Lucas Alexander | alexanderdevelopment.net | 25-Feb-2018 |
| [OpenFaaS on a Kubernetes clusterの仕組み](https://qiita.com/TakanariKo/items/5e3117ea7c3afa948de5) | Takanari Ko | qiita.com | 21-Feb-2018 |
| [VMware Welcomes Alex Ellis](https://blogs.vmware.com/opensource/2018/02/19/vmware-welcomes-alex-ellis/) | VMware team | vmware.com | 19-Feb-2018 |
| [I'm going to be working on OpenFaaS full-time](https://blog.alexellis.io/full-time-openfaas/) | Alex Ellis | alexellis.io | 19-Feb-2018 |
| [Stop installing CLI tools on your build server: CLI-as-a-Function with OpenFaaS](https://medium.com/@burtonr/stop-installing-cli-tools-on-your-build-server-cli-as-a-function-with-openfaas-80dd8d6be611) | Burton Rheutan | medium.com/@burtonr | 12-Feb-2018 |
| [Turn Any CLI into a Function with OpenFaaS](https://blog.alexellis.io/cli-functions-with-openfaas/) | Alex Ellis | alexellis.io | 06-Feb-2018
| [Deploying OpenFaaS on Kubernetes  - Azure AKS](https://medium.com/@ericstoekl/deploying-openfaas-on-kubernetes-azure-aks-4eea99d0743f) | Eric Stoekl | medium.com | 03-Feb-2018 |
| [OpenFaaS with Ruby/MySQL/Ceph](https://github.com/Lallassu/openfaas-demo) | Magnus Persson | cygate.se | 23-Jan-2018 |
| [Installing OpenFaaS on top of Kubernetes on bare-metal with Terraform Scaleway provider and kubeadm](https://stefanprodan.com/2018/kubernetes-scaleway-baremetal-arm-terraform-installer/) | Stefan Prodan | stefanprodan.com | 22-Jan-2018
| [Get storage for your Severless Functions with Minio & Docker](https://blog.alexellis.io/openfaas-storage-for-your-functions/) | Alex Ellis | alexellis.io | 22-Jan-2018
| [ColoriseBot: three months on](https://finnian.io/blog/colorisebot-three-months-on/) | Finnian Anderson | finnian.io | 22-Jan-2018 |
| [Raspberry Pi Cluster with Docker and OpenFaaS (and SETI)](http://6figuredev.com/learning/raspberry-pi-cluster-with-docker-and-openfaas-and-seti/) | John Callaway | 6FigureDev.com | 17-Jan-2018 |
| [Podcast - Episode 022 – OpenFaaS with Docker Captain Alex Ellis](http://6figuredev.com/podcast/episode-022-openfaas-with-docker-captain-alex-ellis/) | The 6 Figure Developer Podcast | 6FigureDev.com | 15-Jan-2018 |
| [Google Home + OpenFaaS: Write your own voice controlled functions](https://medium.com/@burtonr/google-home-openfaas-write-your-own-voice-controlled-functions-11f195398e3f) | Burton Rheutan | medium.com/@burtonr | 06-Jan-2018 |
| [OpenFaaS At The Helm w/ Kubernetes in the DevNet Sandbox](https://blogs.cisco.com/developer/openfaas-at-the-helm-w-kubernetes-in-the-devnet-sandbox) | [Cisco DevNet Sandbox](https://developer.cisco.com/sandbox) | [Cisco DevNet](https://developer.cisco.com) | 2-Jan-2018 |

### 2017

#### Events in 2017
[Back to top](#openfaas-community)

| Event name and description                                          | Speaker      | Location | Date        |
|---------------------------------------------------------------------|--------------|----------|-------------|
| [OpenFaaS @ KubeCon + CloudNativeCon](https://kccncna17.sched.com/event/CU6s/faas-and-furious-0-to-serverless-in-60-seconds-anywhere-alex-ellis-adp) | Alex Ellis | Austin, USA | 07-Dec-2017 |
| [OpenFaaS on Azure AKS with ASP.NET Core website @ AzugFR](https://www.meetup.com/fr-FR/AZUG-FR/events/245186139/) | [Laurent Grangeau](https://twitter.com/laurentgrangeau) | Paris, France | 05-Dec-2017 |
| [OpenFaaS presentation + Demo on Kubernetes @ Paris Serverless Meetup](https://www.meetup.com/fr-FR/Paris-Serverless-Architecture-Meetup/) | [Laurent Grangeau](https://twitter.com/laurentgrangeau) | Paris, France | 14-Nov-2017 |
| [OpenFaaS - Serverless Functions Made Simple @ Docker Cambridge Meetup](https://www.meetup.com/docker-cambridge/events/243206434/) | Alex Ellis | Cambridge, UK | 20-Nov-2017 |
| [OpenFaaS Demo at APIOps Meetup](https://www.meetup.com/APIOps-Tampere/) | [Sakari Hoisko](https://twitter.com/Chaggeus) | Tampere, Finland | 14-Nov-2017 |
| [SecTalks Canberra: OpenFaaS + Security workshop](https://www.meetup.com/SecTalks-Canberra/events/241579721/) | Glenn Grant | Australia | 14-Nov-2017 |
| [Docker Belfast Meetup](https://www.meetup.com/Docker-Belfast/) | John McCabe | Belfast, Northern Ireland | 31-Oct-2017 |
| [Belfast Gophers Meetup](https://www.meetup.com/Belfast-Gophers/) | John McCabe | Belfast, Northern Ireland | 26-Oct-2017 |
| [Introduction to Serverless with OpenFaaS - Moby Summit](https://europe-2017.dockercon.com/moby-summit/) | Alex Ellis | Copenhagen | 19-Oct-2017 |
| [Repainting the Past with Distributed Machine Learning and Docker - DockerCon EU](https://europe-2017.dockercon.com/agenda/#tab_title6) | Finnian Anderson & Oli Callaghan | Copenhagen | 18-Oct-2017 |
| [Zero to Serverless in 60 secs - Dockercon EU](https://europe-2017.dockercon.com/agenda/#tab_title6) | Alex Ellis | Copenhagen | 17-Oct-2017 |
| [Tech Weekly @ UNiDAYS - Alex Young introduces OpenFaaS](https://speakerdeck.com/nullseed/tech-weekly-at-unidays-alex-young-introduces-openfaas) | Alex Young | Nottingham, UK | 12-Oct-2017 |
| [OpenFaaS @ JeffConf Milan](https://milan.jeffconf.com/speakers) | Alex Ellis | Milan, Italy | 29-Sept-2017 |
| [Zero to Serverless in 60 secs - Open Source Summit North America](http://events.linuxfoundation.org/events/open-source-summit-north-america/program/schedule) | Alex Ellis | LA, USA | 11-14 Sept 2017 |
| [FaaS and Furious - CloudNativeLon (Vimeo)](https://skillsmatter.com/skillscasts/10813-faas-and-furious-0-to-serverless-in-60-seconds-anywhere) | Alex Ellis | London, UK | 06 Sept 2017 |
| [From Zero to Serverless with FaaS - Fusion meet-up](https://www.eventbrite.com/e/fusion-meetup-birmingham-june-2017-tickets-35370655583?aff=es2) | Alex Ellis | Birmingham, UK | 20 June 2017 |
| [TechXLR8 - XLR8 your cloud with Docker and Serverless FaaS](https://www.slideshare.net/AlexEllis11/techxlr8-xlr8-your-cloud-with-docker-and-serverless-faas) | Alex Ellis | London, UK | 13-June-2017 |
| [Whaleless!](http://gluecon.com/#agenda) ([slides](https://speakerdeck.com/amirmc/whaleless-doing-serverless-with-docker)) | Amir Chaudhry | GlueCon, CO | 26-May-2017 |
| [MakerShift: Build Your own Serverless functions w/ FaaS](https://www.makershift.io/)| Austin Frey | Harrisburg, PA | 06-May-2017 |
| [Dockercon closing keynote - Cool Hacks](https://blog.docker.com/2017/04/dockercon-2017-mobys-cool-hack-sessions/) | Alex Ellis | Dockercon, Austin | 20-April-2017 |
| [TechLancaster](http://techlancaster.com/meetup)| Austin Frey | Lancaster, PA | 18-April-2017 |
| [ETH Polymese](https://www.polymesse.ch/current-presentations.html)| Brian Christner   | Zürich   | 04-April-2017 |
| [Operationalizing Containers](http://www.rackspace-inform.eu/dispatcher/action?wlmsac=t1ejSDdh1hygBurJJANV)| Brian Christner   | Zürich   | 5-April-2017 |
| [FaaS on Hacker News](https://news.ycombinator.com/item?id=13920588)| Alex Ellis   | Online   | 22-Mar-2017 |


### Blog posts and write-ups 2017
[Back to top](#openfaas-community)

| Blog/repo name and description                                          | Author       | Site     | Date        |
|---------------------------------------------------------------------|--------------|----------|-------------|
| [Podcast - Hanselminutes - Serverless and OpenFaas with Alex Ellis](https://www.hanselminutes.com) | Scott Hanselman | hanselminutes.com | 29-Dec-2017
| [Podcast - How Serverless Technologies Impact Kubernetes](https://thenewstack.io/serverless-technologies-impact-kubernetes/) | The New Stack Makers | thenewstack.io | 28-Dec-2017 |
| [OpenFaaS快速入门指南 (OpenFaaS Quick Start Guide) - Chinese](https://jimmysong.io/posts/openfaas-quick-start/) | Jimmy Song | jimmysong.io | 26-Dec-2017 |
| [OpenFaaS on DCOS](https://medium.com/@realrealbot/openfaas-on-dcos-9d5927f4e725) | Alberto Quario | medium.com/@realrealbot  | 20-Dec-2017
| [Deploying Kubernetes On-Premise with RKE and deploying OpenFaaS on it — Part 1](https://medium.com/@kenfdev/deploying-kubernetes-on-premise-with-rke-and-deploying-openfaas-on-it-part-1-69a35ddfa507) | Ken Fukuyama | medium.com/@kenfdev  | 10-Dec-2017
| [Deploying Kubernetes On-Premise with RKE and deploying OpenFaaS on it — Part 2](https://medium.com/@kenfdev/deploying-kubernetes-on-premise-with-rke-and-deploying-openfaas-on-it-part-2-cc14004e7007) | Ken Fukuyama | medium.com/@kenfdev  | 10-Dec-2017
| [Overview of Functions as a Service with OpenFaas](https://mfarache.github.io/mfarache/Functions-As-A-Service-OpenFaas/) | Mauricio Farache | github.io  | 06-Dec-2017
| [How to secure OpenFaaS with Let's Encrypt and basic auth on Google Kubernetes Engine](https://stefanprodan.com/2017/openfaas-kubernetes-ingress-ssl-gke/) | Stefan Prodan | stefanprodan.com | 04-Dec-2017
| [Running OpenFaaS on GKE - a step by step guide](https://www.weave.works/blog/openfaas-gke) | Stefan Prodan | weave.works | 04-Dec-2017
| [OpenFaaS Meets Perl](https://blog.hex64.com/serverless-open-faas-meets-perl/) | Piotr | hex64.com | 03-Dec-2017
| [Deploying OpenFaaS on Kubernetes - AWS](https://medium.com/@ericstoekl/deploying-openfaas-on-kubernetes-aws-259ec9515e3c) | Eric Stoekl | medium.com/ | 03-Dec-2017
| [Unleash the Artist within you with Docker, Tensorflow and OpenFaaS](http://jmkhael.io/unleash-the-artist-within-tensorflow-and-openfaas/) | Johnny Mkhael  | jmkhael.io | 22-Nov-2017
| [How to Deploy OpenFaaS Serverless PHP Functions](http://blog.gaiterjones.com/how-to-deploy-open-faas-serverless-php-functions/) | Gaiter Jones  | gaiterjones.com | 20-Nov-2017
| [Getting started with OpenFaaS on minikube](https://medium.com/@alexellisuk/getting-started-with-openfaas-on-minikube-634502c7acdf) | Alex Ellis | medium.com | 18-Nov-2017
| [OpenFaaS November Contributor Highlights](https://blog.alexellis.io/openfaas-contributors-highlights/) | Alex Ellis | alexellis.io | 15-Nov-2017 |
| [Contributing to OpenFaaS without writing any code… (yet)](https://hackernoon.com/contributing-to-openfaas-without-writing-any-code-yet-846dd014514f) | Burton Rheutan | hackernoon.com/@burtonr | 11-Nov-2017 |
| [How to start with OpenFaaS](http://panosgeorgiadis.com/blog/2017/11/08/how-to-start-with-openfaas/) | Panos Georgiadis | panosgeorgiadis.com | 08-Nov-2017 |
| [Build a Kubernetes Cluster w/ Raspberry Pi & .NET Core on OpenFaas](https://www.hanselman.com/blog/HowToBuildAKubernetesClusterWithARMRaspberryPiThenRunNETCoreOnOpenFaas.aspx) | Scott Hanselman | hanselman.com | 31-Oct-2017
| [Kubernetes + OpenFaaS + runV + ARM64 Packet.net](https://medium.com/@resouer/kubernetes-openfaas-runv-arm64-packet-net-ea4c7b61c88f) | Harry Zhang | medium.com | 31-Oct-2017 |
| [Serverless Kubernetes home-lab with your Raspberry Pis](https://blog.alexellis.io/serverless-kubernetes-on-raspberry-pi/) | Alex Ellis | alexellis.io | 27-Oct-2017 |
| [Build a Serverless Memes Function with OpenFaaS](https://medium.com/@mlabouardy/build-a-serverless-memes-function-with-openfaas-f4210a53abe8) | Mohamed Labouardy | medium.com/@mlabouardy | 24-Oct-2017 |
| [Repainting the Past with Machine Learning and OpenFaaS](http://olicallaghan.com/post/repainting-the-past-with-machine-learning-and-openfaas) | Oli Callaghan | olicallaghan.com | 23-Oct-2017 |
| [An Introduction to Serverless DevOps with OpenFaaS](https://hackernoon.com/an-introduction-to-serverless-devops-with-openfaas-b978ab0eb2b) | Ken Fukuyama | medium.com/@kenfdev | 22-Oct-2017 |
| [Colourising Video with OpenFaaS Serverless Functions](https://finnian.io/blog/colourising-video-with-openfaas-serverless-functions/) | Finnian Anderson | finnian.io | 21-Oct-2017 |
| [My First Pull Request To OpenFaaS](https://hackernoon.com/my-first-pull-request-to-openfaas-a-major-open-source-project-d0c823790691) | Burton Rheutan | hackernoon.com/@burtonr | 10-Oct-2017 |
| [Deploying a Serverless Youtube-To-Gif Telegram bot with OpenFaaS](https://medium.com/@ericstoekl/deploying-a-serverless-youtube-to-gif-telegram-bot-with-openfaas-2b78d1e9ae6) | Eric Stoekl | medium.com/@ericstoekl | 25-Sep-2017 |
| [A Serverless GraphQL Blog in 60 Seconds with OpenFaaS](https://medium.com/@kenfdev/a-serverless-graphql-blog-in-60-seconds-with-openfaas-aaedd566b1f3) | Ken Fukuyama | medium.com/@kenfdev | 24-Sep-2017 |
| [Create your own DownNotifier with OpenFaaS](http://jmkhael.io/downnotifier-site-pinger/)| Johnny Mkhael | jmkhael.io | 22-Sep-2017|
| [Get OpenFaaS running on Azure Container Service with Kubernetes](https://gist.github.com/danigian/e6097fad36f03c476a69e6c44fde074f) | Daniele Antonio Maggio | twitter.com/danigian | 21-Sep-2017 |
| [OpenFaaS on Azure (Swarm)](https://blogs.technet.microsoft.com/livedevopsinjapan/2017/09/21/openfaas-on-azure-swarm/) | Tsuyoshi Ushio | technet.microsoft.com | 21-Sep-2017 |
| [OpenFaaS on ACS (Kubernetes)](https://blogs.technet.microsoft.com/livedevopsinjapan/2017/09/19/open-faas-on-acs-kubernetes/) | Tsuyoshi Ushio | technet.microsoft.com | 19-Sep-2017 |
| [OpenFaaS on Civo (Kubernetes)](https://www.civo.com/learn/kubernetes-and-openfaas-using-terraform) | Andy Jeffries | civo.com | 18-Sep-2017 |
| [Integrating OpenFaaS and GraphQL](https://medium.com/@kenfdev/integrating-openfaas-and-graphql-experimental-1870bd22f2a) | Ken Fukuyama | medium.com/@kenfdev | 10-Sep-2017 |
| [Morning coffee with the OpenFaaS CLI](https://blog.alexellis.io/quickstart-openfaas-cli/) | Alex Ellis | alexellis.io | 05-Sep-2017 |
| [OpenFaaS on Rancher](https://medium.com/cloud-academy-inc/openfaas-on-rancher-684650cc078e) | Ken Fukuyama | medium.com/@kenfdev | 02-Sep-2017 |
| [OpenFaaS accelerates serverless Java with AfterBurn](https://blog.alexellis.io/openfaas-serverless-acceleration/) | Alex Ellis | alexellis.io | 01-Sep-2017 |
| [OpenFaaS presents to CNCF Serverless workgroup](https://blog.alexellis.io/openfaas-cncf-workgroup/) | Alex Ellis | alexellis.io | 29-Aug-2017 |
| [Your Serverless Raspberry Pi cluster with Docker](https://blog.alexellis.io/your-serverless-raspberry-pi-cluster/) | Alex Ellis | alexellis.io | 20-Aug-2017 |
| [Your first serverless .NET function with OpenFaaS](https://medium.com/@rorpage/your-first-serverless-net-function-with-openfaas-27573017dedb) | Rob Page | medium.com | 19-Aug-2017 |
| [Your first serverless Python function with OpenFaaS](https://blog.alexellis.io/first-faas-python-function/) | Alex Ellis | alexellis.io | 11-Aug-2017 |
| [Introducing Functions as a Service - OpenFaaS](https://blog.alexellis.io/introducing-functions-as-a-service/) | Alex Ellis | alexellis.io | 08-Aug-2017 |
| [Time2Code: Functions as Service and Code as a Function](https://medium.com/@JockDaRock/time2code-functions-as-service-and-code-as-a-function-3d9125fc49fb) | Jock Reed | medium.com/@JockDaRock | 4-Aug-2017|
| [FaaS complete walk-through on Kubernetes (video)](https://www.youtube.com/watch?v=0DbrLsUvaso) | Alex Ellis | youtube.com | 28-Jul-2017 |
| [Kittens vs Tarsiers - an introduction to serverless machine learning](http://jmkhael.io/kittens-vs-tarsiers-an-introduction-to-serverless-machine-learning/) | Johnny Mkhael | jmkhael.io | 24-Jul-2017|
| [Old and new - ASCII art and serverless](http://jmkhael.io/create-a-serverless-ascii-banner-with-faas/) | Johnny Mkhael | jmkhael.io | 29-Jun-2017|
| [Sysdig newsletter](http://go.sysdig.com/webmail/231542/29878667/a2cfc2b92d460d3516188054c1f63e6ecdd7eac458f992f37bbe96acb8d83a78) | Sysdig | sysdig.com | 28-Jun-2017|
| [Ship Serverless FaaS functions with ease](http://blog.alexellis.io/build-and-deploy-with-faas/) | Alex Ellis | alexellis.io | 25-Jun-2017 |
| [Functions as a Service - deploying functions to Docker Swarm via CLI](https://dev.to/developius/functions-as-a-service---deploying-functions-to-docker-swarm-via-a-cli) |  Finnian Anderson | dev.to | 18 May 2017 |
| [Create a Morse Code serverless FaaS function in Quick Basic](http://jmkhael.io/code-a-morse-serverless-function-in-quick-basic-with-faas-and-docker/) | Johnny Mkhael | jmkhael.io | 8 May 2017 |
| ["Serverless"? On my own servers?](http://jmkhael.io/serverless-and-on-my-own-servers/) | Johnny Mkhael | jmkhael.io | 4 May 2017 |
[Ship Serverless Functions to your Docker Swarm with FaaS](https://finnian.io/blog/ship-serverless-functions-to-your-docker-swarm-with-faas/) | Finnian Anderson | finnian.io | 2 May 2017 |
| [Docker + Serverless = FaaS (Functions As A Service)](http://lukeangel.co/cross-platform/docker-servless-faas-functions-as-a-service/) | Luke Angel | lukeangel.co | 28-Apr-2017 |
| [Serverless for private clouds](https://medium.com/@pfernandom/serverless-for-private-clouds-or-managing-the-server-for-a-serverless-app-f9321e45a910) | Pedro Fernando Marquez Soto | medium.com | 30-Apr-2017 |
| [O’Reilly Systems Engineering and Operations newsletter](http://post.oreilly.com/form/oreilly/viewhtml/9z1zlu60fomtoiq41cigtigddvnejvh4vjjfm0i1sh8) | O'Reilly | oreilly.com | 11-Apr-2017 |
| [Test Driving Docker Functions as a Service (FaaS)](https://www.brianchristner.io/test-driving-docker-function-as-a-service-faas/) | Brian Christner | brianchristner.io | 10-Apr-2017 |
| [[cncf-toc] The Cloud-Nativity of Serverless](https://lists.cncf.io/pipermail/cncf-toc/2017-March/000781.html) | CNCF mailing-list | cncf.io | 31-May 2017
| [Docker-awesome curated projects](https://github.com/veggiemonk/awesome-docker#serverless) | Julien Bisconti | Github.com | 23-Mar-2017 |
| [Functions as a Service (FaaS) in French](http://blog.alexellis.io/fonction-service/) | Alex Ellis | alexellis.io | 13-Jan-2017 |
| [Functions as a Service (FaaS)](http://blog.alexellis.io/functions-as-a-service/) | Alex Ellis | alexellis.io | 4-Jan-2017 |

### Repositories
#### Projects, Samples, and Use Cases
[Back to top](#openfaas-community)

You can also find cool projects or submit your own to the [faas-and-furious organisation](https://github.com/faas-and-furious). Contact Alex for details on how.

| Project name and description                                         | Author     | Site      | Date        |
|----------------------------------------------------------------------|------------|-----------|-------------|
| [Faasflow - faas workflow as a function](https://github.com/s8sg/faasflow) | Swarvanu Sengupta | github.com | 11-Nov-2018 |
| [OpenFaaS Crystal Template](https://github.com/TPei/crystal_openfaas) | Thomas Peikert | github.com | 26-Oct-2018 |
| [OpenFaaS Terraform Provider](https://github.com/ewilde/terraform-provider-openfaas) | Edward Wilde | github.com | 12-Oct-2018 |
| [OpenFaaS Swift Template](https://github.com/affix/openfaas-swift-template) | Keiran Smith | github.com | 22-May-2018 |
| [OpenFaaS Code Eval](https://github.com/testdrivenio/openfass-code-eval) | Michael Herman | github.com | 30-Mar-2018 |
| [OpenFaaS RESTful API w/ Node and Postgres](https://github.com/testdrivenio/openfass-node-restful-api) | Michael Herman | github.com | 26-Jan-2018 |
| [OpenFaaS with Ruby/MySQL/Ceph Sample](https://github.com/Lallassu/openfaas-demo) | Magnus Persson | [lallassu/openfaas-demo](https://github.com/Lallassu/openfaas-demo) | 23-Jan-2018
| [sqlrest: API Proxy to connect to a SQL database from a function](https://github.com/BurtonR/sqlrest) | Burton Rheutan | github.com | 21-Nov-2017
| [Kafka-connector for OpenFaaS](https://github.com/ucalgary/ftrigger) | King Chung Huang | [ucalgary/ftrigger](https://github.com/ucalgary/ftrigger) | 13-Oct-2017
| [Azure Storage and Azure Service Bus Queues Trigger for OpenFaaS on Kubernetes](https://github.com/danigian/Azure-Queues-Trigger-OpenFaaS) | Daniele Antonio Maggio | github.com/danigian/Azure-Queues-Trigger-OpenFaaS | 24-Sep-2017 |
| [faas-pipeline](https://github.com/getlazy/faas-pipeline) - OpenFaaS function composition | Ivan Erceg | github.com | 23-Sep-2017 |
| [Github - Reverse geocoding](https://github.com/lucj/faas-reverse-geocoding) | Luc Juggery | github.com | 14-Sep-2017 |
| [QR Code](https://github.com/faas-and-furious/qrcode) | John McCabe | github.com | 18-Aug-2017 |
| [Img2ANSI - GIF/PNG/JPG to ANSI art converter](https://github.com/johnmccabe/faas-img2ansi/) | John McCabe | github.com | 10-Aug-2017 |
| [FaaS-netes - Kubernetes backend for FaaS](https://github.com/openfaas/faas-netes) | Alex Ellis | github.com | 25-Jul-2017 |
| [Twitter, Elastic Search and Alexa stack of functions](https://github.com/alexellis/journey-expert/tree/master/tweetstash) | Alex Ellis | github.com | 24-Apr-2017 |
| [Github - URL Shortener](https://github.com/developius/faas-node-url-shortener) | Finnian Anderson | [developius/faas-node-url-shortener](https://github.com/developius/faas-node-url-shortener) | 19-Apr-2017 |
| [Github - fibonacci numbers up to N](https://github.com/developius/faas-python-fib/) | Finnian Anderson | [developius/faas-python-fib/](https://github.com/developius/faas-python-fib/) | 19-Apr-2017 |
| [Github - ascii art generator](https://github.com/developius/faas-node-ascii) | Finnian Anderson | [developius/faas-node-ascii](https://github.com/developius/faas-node-ascii) | 17-Apr-2017 |
| [Dockercon FaaS demos including Alexa/Github](https://github.com/alexellis/faas-dockercon) | Alex Ellis | github.com | 14-Apr-2017 |
| [Docker Birthday voting app ported to FaaS](https://github.com/alexellis/faas-example-voting-app) | Alex Ellis | github.com | 14-Apr-2017 |
| [Github-Airtable Bug Tracker integration](https://github.com/aafrey/faas-demo) | Austin Frey | github.com/aafrey/faas-demo | 10-Apr-2017 |
