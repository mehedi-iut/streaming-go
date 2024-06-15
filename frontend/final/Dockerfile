FROM nginx
LABEL authors="mehedi"

COPY .. /usr/share/nginx/html/

EXPOSE 80

ENTRYPOINT ["nginx", "-g", "daemon-off"]