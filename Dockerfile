FROM rust:latest as builder

# create project, note: it will add the hello world stuff
RUN USER=root cargo new --bin chord-paper-be
WORKDIR ./chord-paper-be

# Copy only the Cargo.toml to working directory, so at this point it will be an empty project
# (bar hello world in main.rs) and all the dependencies.  The goal here is to build the deps
# first.
COPY ./Cargo.toml ./Cargo.toml
RUN cargo build --release --all-features

# after building, remove the src directory (basically the hello world main.rs), and add back
# all the real code in the src folder.
RUN rm src/*.rs
ADD . ./

# remove the hello world main.rs binaries, and now build the actual code
RUN rm ./target/release/deps/chord_paper_be*
RUN cargo build --release

# now for the actual image that runs the program, use debian image:
FROM debian:buster-slim
ARG APP=/usr/src/app

# install time zone package and ca-certs, for best practice (logs, and in case https used 
# in the future)
RUN apt-get update \
    && apt-get install -y ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

# expose the port
EXPOSE 5000 

# best practice: create a new user and group, make a working directory for that user
ENV TZ=America/Vancouver \
    APP_USER=appuser
RUN groupadd $APP_USER \
    && useradd -g $APP_USER $APP_USER \
    && mkdir -p ${APP}

# copy the binaries from the previous stage to the working directory of this stage,
# give permissions, run the program
COPY --from=builder /chord-paper-be/target/release/chord-paper-be ${APP}/chord-paper-be
RUN chown -R $APP_USER:$APP_USER ${APP}

USER $APP_USER
WORKDIR ${APP}

CMD ["./chord-paper-be"]

