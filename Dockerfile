FROM ubuntu:14.04
MAINTAINER xujie xujieasd@gmail.com
RUN mkdir hkrest
RUN mkdir -p /hkrest/HCNetSDKCom
COPY build/HCNetSDKCom/*.so /hkrest/HCNetSDKCom/
COPY build/main /hkrest/
COPY build/*.so /hkrest/
COPY entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]