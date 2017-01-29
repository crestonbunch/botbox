import { observable, autorun } from "mobx";
import "whatwg-fetch"

const MAX_NAME_LENGTH = 20;
const MIN_PASSWORD_LENGTH = 6;

/**
 * A data store for a registration form.
 */
export class RegisterStore {
  @observable name: string = "";
  @observable email: string = "";
  @observable password: string = "";
  @observable recaptcha: string = "";

  @observable nameError: string = "";
  @observable emailError: string = "";
  @observable passwordError: string = "";
  @observable recaptchaError: string = "";
  @observable serverError: string = "";

  // tracks if an asynchronous operation is running
  // that should disable the form until it is done.
  @observable busy: boolean = false;

  // used for checking if an asynchronous validation
  // process is underway, to stop submitting the
  // registration until it is finished
  @observable validating: boolean = false;

  // if the registration was successfully submitted, this
  // will be set to true.
  @observable success: boolean = false;

  validateName = autorun(function () {
    if (this.name.length > MAX_NAME_LENGTH) {
      this.nameError = "Name must be no more than 20 characters.";
    } else {
      this.nameError = "";
    }
  }.bind(this));

  validateEmail = autorun(function () {
    if (this.email != "") {
      this.validating = true;
      fetch('/api/email/' + this.email).
        then((response: Response) => {
          if (response.status == 200) {
            this.emailError = "That email is already in use!";
          } else {
            this.emailError = "";
          }
          this.validating = false;
        }).catch(() => {
          this.validating = false;
        });
    }
  }.bind(this));

  validatePassword = autorun(function () {
    if (this.password != "") {
      this.passwordError = "";
    }
  }.bind(this));

  validateRecaptcha = autorun(function () {
    if (this.recatcha == "") {
      this.recaptchaError = "Please complete the captcha."
    } else {
      this.recaptchaError = "";
    }
  }.bind(this))

  doRegister() {
    if (this.validating) return;
    this.serverError = "";

    let valid = true;
    if (this.name == "") {
      this.nameError = "Please enter a display name.";
      valid = false;
    }
    if (this.email == "") {
      this.emailError = "Please enter an email.";
      valid = false;
    } else if (this.email.indexOf('@') < 0) {
      this.emailError = "Please provide a valid email.";
      valid = false;
    }
    if (this.password == "") {
      this.passwordError = "Please enter a password.";
      valid = false;
    } else if (this.password.length < MIN_PASSWORD_LENGTH) {
      this.passwordError = "Password must be at least 6 characters."
      valid = false;
    }
    if (this.recaptcha == "") {
      this.recaptchaError = "Please answer the captcha."
      valid = false;
    }
    if (!valid) return;

    this.busy = true;

    let request = {
      name: this.name,
      email: this.email,
      password: this.password,
      captcha: this.recaptcha,
    }

    fetch('/api/user', {
      method: 'POST',
      body: JSON.stringify(request)
    }).then((response: Response) => {
      if (response.status == 200) {
        this.success = true;

        // User account was created, so send an email.
        fetch('/api/email/verify', {
          method: 'POST',
          body: JSON.stringify({email: this.email})
        }).then((response: Response) => {
          // There was an error with the email sending service.
          if (response.status != 200) {
            response.text().then((message: string) => {
              this.serverError = message;
            });
          }
        });
      } else {
        response.text().then((message: string) => {
          this.serverError = message;
        });
      }
      this.busy = false;
    });
  }
}