import { observable, autorun } from "mobx";
import { UserStore } from "../domain/user";
import "whatwg-fetch"

/**
 * An expected response from the server after authenticating.
 */
interface LoginResponse {
  user: number;
  secret: string;
  expiration: string;
}

interface UserResponse {
  name: string;
  joined: Date;
  permission_set: string;
  permissions: string[];
}

/*
* A data store for a login form.
*/
export class LoginStore {

  @observable email: string = "";
  @observable password: string = "";

  @observable error: string = "";

  // tracks if an asynchronous operation is running
  // that should disable the form until it is done.
  @observable busy: boolean = false;

  userStore: UserStore;

  constructor(userStore: UserStore) {
    this.userStore = userStore;
  }

  /**
   * Attempt to authenticate the user and fill the user store.
   */
  doLogin() {
    this.busy = true;

    fetch('/api/session', {
      method: "POST",
      body: JSON.stringify({ email: this.email, password: this.password })
    }).then(function(response: Response) {
      if (response.status == 200) {
        this.userStore.session = {};
        response.json().then(function(value: LoginResponse) {
          this.userStore.session.id = value.user;
          this.userStore.session.secret = value.secret;
          this.userStore.session.expiration = Date.parse(value.expiration);
          return fetch('/api/user/id/' + String(value.user));
        }.bind(this)).then(function(response: Response) {
          if (response.status == 200) {
            response.json().then(function(value: UserResponse) {
              this.userStore.session.name = value.name;
              this.userStore.session.permissions = value.permissions;
              this.userStore.session.permissionSet = value.permission_set;
              this.userStore.loggedIn = true;
            }.bind(this));
          } else {
            response.text().then(function(value: string) {
              this.error = value;
            }.bind(this));
          }
        }.bind(this));
      } else {
        response.text().then(function(value: string) {
          this.error = value;
        }.bind(this));
      }
      this.busy = false;
    }.bind(this));
  }

  /**
   * De-authorize the user's session and clear the user store.
   */
  doLogout() {

  }

}
