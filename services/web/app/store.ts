import { observable, autorun, action } from "mobx";
import "whatwg-fetch"

const LS_USER_KEY = "LS_USER_KEY";

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

/**
 * A data store that holds information about the currently logged-in user.
 */
export class Store {

  // A boolean set if the user is logged in or not.
  @observable loggedIn: boolean = false;

  // Information about the user's session
  @observable session: SessionData;

  // Restore session from local storage.
  restoreSession = autorun(function() {
    let savedSession = localStorage.getItem(LS_USER_KEY);
    if (savedSession != null && !this.loggedIn) {
      this.session = JSON.parse(savedSession) as SessionData;
      this.loggedIn = true;
    }
  }.bind(this))

  // Save a session to localstorage
  saveSession = autorun(function() {
    if (this.loggedIn) {
      localStorage.setItem(LS_USER_KEY, JSON.stringify(this.session));
    } else {
      localStorage.removeItem(LS_USER_KEY);
    }
  }.bind(this))

  // Perform the login action on the state.
  @action login(session: SessionData) {
    this.session = session;
    this.loggedIn = true;
  }

}

