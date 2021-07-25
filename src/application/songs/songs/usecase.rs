use super::dynamodb;
use super::entity;
use crate::application::concerns::user_validation;
use snafu::Snafu;

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
    user_validation: user_validation::UserValidation,
    datastore: dynamodb::DynamoDB,
}

impl Usecase {
    pub fn new(
        user_validation: user_validation::UserValidation,
        songs_datastore: dynamodb::DynamoDB,
    ) -> Usecase {
        Usecase {
            user_validation: user_validation,
            datastore: songs_datastore,
        }
    }

    pub async fn get_song(&self, id: &str) -> Result<entity::Song, Error> {
        if !entity::Song::is_valid_id(id) {
            // if it's not a UUID we won't find it in the datastore
            // just short circuit and don't hit the DB
            return Err(Error::NotFoundError { id: id.to_string() });
        }

        let get_song_result = self.datastore.get_song(id).await;
        get_song_result.map_err(|err| map_datastore_error(err, id))
    }

    pub async fn create_song(
        &self,
        user_id_token: &str,
        mut song: entity::Song,
    ) -> Result<entity::Song, Error> {
        if !song.is_new() {
            return Err(Error::ExistingSongError);
        }

        self.user_validation
            .verify_owner(user_id_token, &song.summary)
            .map_err(map_user_validation_error)?;

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
        if song_id.is_empty() {
            return Err(Error::NotFoundError { id: "".to_string() });
        }

        if song_id != song.summary.id {
            return Err(Error::WrongIDError {
                song_id_1: song_id.to_string(),
                song_id_2: song.summary.id.to_string(),
            });
        }

        //TODO - actually fetch the song instead, potential vulnerability here
        let validation_result = self
            .user_validation
            .verify_owner(user_id_token, &song.summary);

        validation_result.map_err(map_user_validation_error)?;

        self.protect_overwrite_song(&song).await?;

        song.set_saved_time();

        match self.datastore.update_song(&song).await {
            Ok(()) => Ok(song),
            Err(err) => Err(map_datastore_error(err, &song.summary.id)),
        }
    }

    pub async fn delete_song(&self, user_id_token: &str, song_id: &str) -> Result<(), Error> {
        if song_id.is_empty() {
            return Err(Error::NotFoundError { id: "".to_string() });
        }

        let song = self.get_song(song_id).await?;

        if song_id != song.summary.id {
            return Err(Error::WrongIDError {
                song_id_1: song_id.to_string(),
                song_id_2: song.summary.id.to_string(),
            });
        }

        let validation_result = self
            .user_validation
            .verify_owner(user_id_token, &song.summary);

        validation_result.map_err(map_user_validation_error)?;

        match self.datastore.delete_song(song_id).await {
            Ok(()) => Ok(()),
            Err(err) => Err(map_datastore_error(err, song_id)),
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
                    return Err(Error::OverwriteError);
                }
            }
            None => {} // really unexpected....but don't block on this I guess
        }

        Ok(())
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

fn map_user_validation_error(err: user_validation::Error) -> Error {
    match err {
        user_validation::Error::GoogleVerificationError { source } => {
            Error::GoogleVerificationError { source }
        }
        user_validation::Error::WrongOwnerError => Error::WrongOwnerError,
    }
}
