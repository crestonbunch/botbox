export interface SessionData {
  // The user's serial from the database.
  id: number;

  // The display name for the user.
  name: string;

  // The user's email address.
  email: string;

  // Name of the user's permission set.
  permissionSet: string;

  // List of permissions the user has.
  permissions: string[];

  // The user's session secret.
  secret: string;

  // The time the session expires.
  expiration: Date;
}

export interface Notification {
    id: number,
    type: string,
    parameters: any | null,
    issued: Date
    read: Date | null,
    dismissed: Date | null,
}

