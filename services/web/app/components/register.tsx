import * as React from "react";
import "whatwg-fetch"
import { Label, Message, Input, Icon, Form, Modal, Button } from "semantic-ui-react";
import { Recaptcha } from "./recaptcha";
import { Settings } from "../settings";

const PASSWORD_LENGTH = 6;
const USERNAME_LENGTH = 20;

export interface RegisterProps {
  trigger: JSX.Element;
}

export interface RegisterState {
  open: boolean;
  pending: boolean;
  success: boolean;
  error: boolean;
}

export class Register extends React.Component<RegisterProps, RegisterState> {

  private recaptcha: Recaptcha;

  private requestBody = {
    username: "",
    email: "",
    password: "",
    passconf: "",
    captcha: "",
  };

  private errors: {
    serverError?: string;
    usernameError?: string;
    emailError?: string;
    passwordError?: string;
    passconfError?: string;
    captchaError?: string;
  } = {};

  constructor(props: RegisterProps) {
    super(props);
    this.state = {
      open: false,
      pending: false,
      success: false,
      error: false,
    };
  }

  updateUsername(e: Event) {
    this.requestBody.username = (e.target as any).value;

    if (this.requestBody.username.length > USERNAME_LENGTH) {
      this.errors.usernameError =
        "Username must be no more than 20 characters.";
      this.forceUpdate();
    } else if (this.errors.usernameError != undefined) {
      this.errors.usernameError = undefined;
      this.forceUpdate();
    }

    if (this.requestBody.username.length > 0) {
      fetch('/api/account/exists/username/' + this.requestBody.username).
        then(function(response: Response) {
          if (response.status == 200) {
            response.text().then(function(result: string) {
              if (result == "true") {
                this.errors.usernameError = "That username is already taken!"
                this.forceUpdate();
              }
            }.bind(this));
          }
        }.bind(this))
    }
  }

  updatePassword(e: Event) {
    this.requestBody.password = (e.target as any).value;
  }

  updatePassconf(e: Event) {
    this.requestBody.passconf = (e.target as any).value;
  }

  updateEmail(e: Event) {
    this.requestBody.email = (e.target as any).value;

    if (this.requestBody.email.length > 0) {
      fetch('/api/account/exists/email/' + this.requestBody.email).
        then(function(response: Response) {
          if (response.status == 200) {
            response.text().then(function(result: string) {
              if (result == "true") {
                this.errors.emailError = "That email is already is use!"
                this.forceUpdate();
              }
            }.bind(this));
          }
        }.bind(this))
    }
  }

  updateCaptcha(response: string) {
    this.requestBody.captcha = response;
  }

  doRegister() {
    if (this.validate()) {
      this.setState({
        open: true,
        pending: true,
        success: false,
        error: false,
      })

      fetch('/api/account/new', {
        method: 'POST',
        body: JSON.stringify(this.requestBody)
      }).then(function(response: Response) {
        this.recaptcha.reset();
        if (response.status == 200) {
          this.setState({
            open: true,
            pending: false,
            success: true,
            error: false,
          });
        } else {
          response.text().then(function(message: string) {
            this.errors.serverError = message;
            this.setState({
              open: true,
              pending: false,
              success: false,
              error: true,
            });
          }.bind(this));
        }
      }.bind(this));
    } else {
      this.setState({
        open: true,
        pending: false,
        success: false,
        error: true,
      });
    }
  }

  validate() : boolean {
    this.errors = {};

    if (this.requestBody.username.length == 0) {
      this.errors.usernameError = "Please provide a username.";
    }
    if (this.requestBody.email.length == 0) {
      this.errors.emailError = "Please provide an email.";
    } else if (this.requestBody.email.indexOf('@') == -1) {
      this.errors.emailError = "Please provide a valid email.";
    }
    if (this.requestBody.password.length == 0) {
      this.errors.passwordError = "Please provide a password.";
    } else if (this.requestBody.password.length < PASSWORD_LENGTH) {
      this.errors.passwordError = "Passwords must be at least 6 characters.";
    } else if (this.requestBody.passconf != this.requestBody.password) {
      this.errors.passconfError = "Passwords must match.";
    }
    if (this.requestBody.captcha.length == 0) {
      this.errors.captchaError = "Please prove you are not a robot.";
    }

    let valid = (JSON.stringify(this.errors) == "{}")

    return valid
  }

  open() {
    this.requestBody = {
      username: "",
      email: "",
      password: "",
      passconf: "",
      captcha: "",
    };

    this.errors = {};

    this.setState({
      open: true,
      pending: false,
      success: false,
      error: false,
    });
  }

  close() {
    this.setState({
      open: false,
      pending: false,
      success: false,
      error: false,
    })
  }

  render() {
    let modal = (<Modal size="small" trigger={this.props.trigger} open={this.state.open} onOpen={this.open.bind(this)}
      onClose={this.close.bind(this)}>
      <Modal.Header><Icon name="add user" /> Register</Modal.Header>
      <Modal.Content>
        {(this.state.success) ? (
          <Message success>
            You have been successfully registered! Please check your email to finish verifying your account.
          </Message>
        ) : (
          <Form error={this.state.error} loading={this.state.pending}>
            {this.errors.serverError ? <Message error>{this.errors.serverError}</Message> : null}
            <Form.Field required error={this.errors.usernameError != undefined}>
              <label>Username</label>
              {this.errors.usernameError ? <Label basic pointing="below" color="red"><Icon name="warning sign" />{this.errors.usernameError}</Label> : null}
              <Input icon='users' onChange={this.updateUsername.bind(this)} iconPosition='left' name="username"
                placeholder="Enter a unique username" />
            </Form.Field>
            <Form.Field required  error={this.errors.emailError != undefined}>
              <label>Email</label>
              {this.errors.emailError ? <Label basic pointing="below" color="red">{this.errors.emailError}</Label> : null}
              <Input icon='mail' onChange={this.updateEmail.bind(this)} iconPosition='left' name="email"
                placeholder="email@example.com" />
            </Form.Field>
            <Form.Field required error={this.errors.passwordError != undefined}>
              <label>Password</label>
              {this.errors.passwordError ? <Label basic pointing="below" color="red">{this.errors.passwordError}</Label> : null}
              <Input icon='lock' onChange={this.updatePassword.bind(this)} iconPosition='left' name="password"
                type="password" placeholder="Make it secure" />
            </Form.Field>
            <Form.Field required error={this.errors.passconfError != undefined}>
              <label>Confirm Password</label>
              {this.errors.passconfError ? <Label basic pointing="below" color="red">{this.errors.passconfError}</Label> : null}
              <Input icon='check' onChange={this.updatePassconf.bind(this)} iconPosition='left' name="confirm"
                type="password" placeholder="Retype your password" />
            </Form.Field>
            <Form.Field required error={this.errors.captchaError != undefined}>
              <label>Singularity check</label>
              {this.errors.captchaError ? <Label basic pointing="below" color="red">{this.errors.captchaError}</Label> : null}
              <Recaptcha elementID="registerCaptcha" sitekey={Settings.RECAPTCHA_KEY}
                callback={this.updateCaptcha.bind(this)} ref={e => this.recaptcha = e} />
            </Form.Field>
          </Form>
        )}
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={this.close.bind(this)}>Close</Button>
        {(this.state.success) ? (
          null
        ) : (
          <Button positive disabled={this.state.pending} onClick={this.doRegister.bind(this)}>Join</Button>
        )}
      </Modal.Actions>
    </Modal>);

    return modal;
  }
}

