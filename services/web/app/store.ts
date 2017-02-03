import { observable, autorun, action } from "mobx";
import "whatwg-fetch"
import { Api } from "./api";
import { SessionData, Notification } from "./models";

const LS_USER_KEY = "LS_USER_KEY";

/**
 * A data store that holds information about the currently logged-in user.
 */
export class Store {

  // A boolean set if the user is logged in or not.
  @observable loggedIn: boolean = false;

  // Information about the user's session
  @observable session: SessionData;

  // Notifications fetch from the user backend
  @observable notifications: Notification[] = [];

  // Restore session from local storage.
  restoreSession = autorun(function () {
    let savedSession = localStorage.getItem(LS_USER_KEY);
    if (savedSession != null && !this.loggedIn) {
      this.login(JSON.parse(savedSession) as SessionData);
    }
  }.bind(this))

  // Save a session to localstorage
  saveSession = autorun(function () {
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
    // fetch latest user notifications
    Api.notifications(this.session.secret).then((n: Notification[]) => {
      this.notifications = n;
    });
  }

}

