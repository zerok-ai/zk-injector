FROM alpine
WORKDIR /opt/zerok
COPY resources/ .
RUN chmod +x ./zerok-agent.sh
CMD ["echo", "Delivering the Node JS Agent Install script. Copy the script run-with-agent.sh from /opt/zerok to the application container and override the application entry point to call the script", "; sleep 3"]