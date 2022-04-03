FROM debian
RUN apt -q update && apt install -qy curl parallel