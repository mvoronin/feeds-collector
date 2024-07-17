FROM node:20

ARG UID=1010
ARG GID=1010

RUN groupadd -g $GID nodegroup && \
    useradd -r -u $UID -g nodegroup nodeuser

WORKDIR /app

USER nodeuser

CMD ["npm", "run", "dev"]

#CMD ["npm", "run", "preview"]
#CMD ["/bin/bash"]
