use super::dynamodb;
use super::entity;
use crate::application::concerns::google_verification;
use crate::application::users;
use snafu::Snafu;
use std::str::FromStr;

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("Failed Google token verification: {}", source))]
    GoogleVerificationError { source: google_signin::Error },

    #[snafu(display("Song cannot be created if it already has an ID"))]
    ExistingSongError,

    #[snafu(display(
        "The song ID to be modified is not equal to ID inside the song data: {} and {}",
        song_id_1,
        song_id_2
    ))]
    WrongIDError {
        song_id_1: String,
        song_id_2: String,
    },

    #[snafu(display(
    "This song was previously saved at a more recent time - overwriting is will cause data loss.",
    ))]
    OverwriteError,

    #[snafu(display("Song owner must be equal to the user persisting the song"))]
    WrongOwnerError,

    #[snafu(display("Data store failed: {}", source))]
    DatastoreError { source: dynamodb::Error },

    #[snafu(display("Song ID not found: {}", id))]
    NotFoundError { id: String },
}

#[derive(Clone)]
pub struct Usecase {
    google_verification: google_verification::GoogleVerification,
    datastore: dynamodb::DynamoDB,
}

impl Usecase {
    pub fn new(
        google_verification: google_verification::GoogleVerification,
        songs_datastore: dynamodb::DynamoDB,
    ) -> Usecase {
        Usecase {
            google_verification: google_verification,
            datastore: songs_datastore,
        }
    }

    pub async fn get_song(&self, id: &str) -> Result<entity::Song, Error> {
        validate_song_id(id)?;

        let get_song_result = self.datastore.get_song(id).await;

        match get_song_result {
            Ok(song) => Ok(song),
            Err(err) => Err(map_datastore_error(err, id)),
        }
    }

    pub async fn create_song(
        &self,
        user_id_token: &str,
        mut song: entity::Song,
    ) -> Result<entity::Song, Error> {
        let user = self.verify_user(user_id_token)?;

        if !song.is_new() {
            return Err(Error::ExistingSongError);
        }

        if song.summary.owner != user.id {
            return Err(Error::WrongOwnerError);
        }

        song.create_id();
        song.set_saved_time();

        match self.datastore.create_song(&song).await {
            Ok(()) => Ok(song),
            Err(err) => Err(map_datastore_error(err, &song.summary.id)),
        }
    }

    pub async fn update_song(
        &self,
        user_id_token: &str,
        song_id: &str,
        mut song: entity::Song,
    ) -> Result<entity::Song, Error> {
        let user = self.verify_user(user_id_token)?;

        if song_id.is_empty() {
            return Err(Error::NotFoundError { id: "".to_string() });
        }

        if song_id != song.summary.id {
            return Err(Error::WrongIDError {
                song_id_1: song_id.to_string(),
                song_id_2: song.summary.id.to_string(),
            });
        }

        if song.summary.owner != user.id {
            return Err(Error::WrongOwnerError);
        }

        self.protect_overwrite_song(&song).await?;

        song.set_saved_time();

        match self.datastore.update_song(&song).await {
            Ok(()) => Ok(song),
            Err(err) => Err(map_datastore_error(err, &song.summary.id)),
        }
    }

    fn verify_user(&self, user_id_token: &str) -> Result<users::entity::User, Error> {
        match self.google_verification.verify(user_id_token) {
            Ok(user) => Ok(user),
            Err(err) => Err(Error::GoogleVerificationError { source: err }),
        }
    }

    async fn protect_overwrite_song(&self, song_to_update: &entity::Song) -> Result<(), Error> {
        let next_last_saved_at = song_to_update
            .summary
            .last_saved_at
            .ok_or(Error::OverwriteError)?;

        let curr_song = self.get_song(&song_to_update.summary.id).await?;
        match curr_song.summary.last_saved_at {
            Some(curr_last_saved_at) => {
                // prevent overwriting a more recent save if the last saved at timestamp is greater than the current one
                // example:
                // A --> B
                //   \----->C
                //
                // Suppose I open the song at time A on two computers, make edits on both somewhat absentmindedly
                // I first save at time B - the payload for the song for the last saved at would be A
                // presume this succeeds, the server copy is now from time B and its last saved at time is also B
                //
                // Now I save another copy at time C, but the predecessor of my copy at time C was from A
                // If I just go with last write wins, then all changes from the copy at time B would be overwritten
                // So by comparing the timestamp at the save at C (A vs B), the save will fail to protect overwriting data.
                //
                // The user can then refetch the copy at time B and copy over their changes from time C (manual merge)
                // and form a copy that will be accepted by the server:
                //
                // A --> B----->B+C
                //   \----->C---/
                if curr_last_saved_at.gt(&next_last_saved_at) {
                    log::info!("Overwriting error encountered");
                    log::info!("Current saved at: {}", curr_last_saved_at.to_rfc3339());
                    log::info!("Next saved at: {}", next_last_saved_at.to_rfc3339());

                    return Err(Error::OverwriteError);
                }
            }
            None => {} // really unexpected....but don't block on this I guess
        }

        Ok(())
    }
}

fn validate_song_id(id: &str) -> Result<(), Error> {
    match uuid::Uuid::from_str(id) {
        Ok(_) => Ok(()),
        // if it's not a UUID we won't find it in the datastore
        // just short circuit and don't hit the DB
        Err(_) => Err(Error::NotFoundError { id: id.to_string() }),
    }
}

fn map_datastore_error(err: dynamodb::Error, song_id: &str) -> Error {
    match err {
        dynamodb::Error::NotFoundError => Error::NotFoundError {
            id: song_id.to_string(),
        },
        dynamodb::Error::GenericDynamoError { .. }
        | dynamodb::Error::MalformedDataError { .. }
        | dynamodb::Error::SongSerializationError { .. } => Error::DatastoreError { source: err },
    }
}
