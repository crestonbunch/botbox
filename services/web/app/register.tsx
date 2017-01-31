import * as React from "react";
import { browserHistory } from 'react-router'
import { observer } from "mobx-react";
import { observable, autorun, action } from "mobx";
import { Input, Divider, Grid, Image, Message, Icon, Form, Button, Container, Segment } from "semantic-ui-react";
import { Api } from './api';
import { Recaptcha } from "./components/recaptcha";
import { FieldError } from "./components/forms";
import { Settings } from "./settings";
import { Store, SessionData } from "./store"

export interface RegisterProps {
  store: Store;
}

export interface RegisterState {
}

@observer
export class Register extends React.Component<RegisterProps, RegisterState> {

  @observable name: string = "";
  @observable email: string = "";
  @observable password: string = "";
  @observable captcha: string = "";

  @observable nameError: string = "";
  @observable emailError: string = "";
  @observable passwordError: string = "";
  @observable captchaError: string = "";
  @observable serverError: string = "";

  @observable success: boolean = false;
  @observable busy: boolean = false;
  @observable validating: boolean = false;

  checkNameLength = autorun(() => {
    if (this.name.length > Api.MAX_NAME_LENGTH) {
      this.nameError = "Name must be no more than 20 characters.";
    } else {
      this.nameError = "";
    }
  });

  checkEmailExists = autorun(() => {
    if (this.email == "") return;

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
  });

  clearPasswordError = autorun(() => {
    if (this.password.length > Api.MIN_PASSWORD_LENGTH) {
      this.passwordError = "";
    }
  });

  clearCaptchaError = autorun(() => {
    if (this.captcha != "") {
      this.captchaError = "";
    }
  });

  @action doRegister() {
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
    } else if (this.password.length < Api.MIN_PASSWORD_LENGTH) {
      this.passwordError = "Password must be at least 6 characters."
      valid = false;
    }
    if (this.captcha == "") {
      this.captchaError = "Please answer the captcha."
      valid = false;
    }
    if (!valid) return;

    this.busy = true;

    Api.register(this.name, this.email, this.password, this.captcha)
      .then((result: void) => {
      }).catch((reason: any) => {
        this.serverError = reason as string;
        this.busy = false;
      }).then(() => {
        return Api.login(this.email, this.password);
      }).catch((reason: any) => {
        this.serverError = reason as string;
        this.busy = false;
      }).then((session: SessionData) => {
        this.busy = false;
        this.props.store.login(session);
        browserHistory.push("/");
      });
  }

  constructor(props: RegisterProps) {
    super(props);
  }

  render() {
    const successMsg = (this.success === true) ? (
      <Message success>
        You have been successfully registered!
        Please check your email to finish verifying your account.
      </Message>
    ) : null;

    const errMsg = (this.serverError !== "") ? (
      <Message error>{this.serverError}</Message>
    ) : null;

    const form = (this.success === false) ? (
      <Form error={this.serverError != ""} loading={this.busy}>
        <Form.Field error={this.nameError != ""}>
          <FieldError error={this.nameError} />
          <Input icon="user" iconPosition="left"
            placeholder="Enter a display name."
            onChange={(_, val) => this.name = val.value}
            type="text" />
        </Form.Field>
        <Form.Field>
          <FieldError error={this.emailError} />
          <Input icon="mail" iconPosition="left"
            placeholder="Enter a valid email."
            onChange={(_, val) => this.email = val.value}
            type="text" />
        </Form.Field>
        <Form.Field>
          <FieldError error={this.passwordError} />
          <Input icon="lock" iconPosition="left"
            placeholder="Enter a good password."
            onChange={(_, val) => this.password = val.value}
            type="password" />
        </Form.Field>
        <Form.Field required error={this.captchaError != ""}>
          <FieldError error={this.captchaError} />
          <Recaptcha sitekey={Settings.RECAPTCHA_KEY}
            onChange={(val) => this.captcha = val} />
        </Form.Field>
        <Divider hidden />
        <Button primary
          disabled={this.validating}
          onClick={(e) => { e.preventDefault(); this.doRegister() }}>
          Join
        </Button>
      </Form>
    ) : null;

    return (
      <Segment basic vertical>
        <Image size="medium" src="assets/botbox-masthead.svg" centered />
        <Container>
          <Grid padded centered>
            <Grid.Column largeScreen={5} computer={8} tablet={10} mobile={14}>
              <Segment>
                {successMsg}
                {errMsg}
                {form}
              </Segment>
            </Grid.Column>
          </Grid>
        </Container>
      </Segment>
    )
  }
}