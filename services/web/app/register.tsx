import * as React from "react";
import "whatwg-fetch"
import { Divider, Grid, Image, Message, Input, Icon, Form, Button, Container, Segment } from "semantic-ui-react";
import { Recaptcha } from "./components/register/recaptcha";
import { Name } from "./components/register/name";
import { Email } from "./components/register/email";
import { Password } from "./components/register/password";

const PASSWORD_LENGTH = 6;

export interface RegisterProps {
}

export interface RegisterState {
  open?: boolean;
  pending?: boolean;
  success?: boolean;
  error?: string;
}

export class Register extends React.Component<RegisterProps, RegisterState> {

  private name: Name;
  private email: Email;
  private password: Password;
  private recaptcha: Recaptcha;

  constructor(props: RegisterProps) {
    super(props);
    this.state = {
      open: false,
      pending: false,
      success: false
    };
  }

  doRegister() {
    console.log("REGISTERING");

    this.setState({
      pending: true,
      success: false,
      error: undefined,
    });

    this.validate().then(function (valid: boolean) {
      if (valid) {

        let request = {
          name: this.name.getValue(),
          email: this.email.getValue(),
          password: this.password.getValue(),
          captcha: this.recaptcha.getValue(),
        }

        fetch('/api/user', {
          method: 'POST',
          body: JSON.stringify(request)
        }).then(function (response: Response) {
          this.recaptcha.reset();
          if (response.status == 200) {
            this.setState({
              pending: false,
              success: true,
            });
          } else {
            response.text().then(function (message: string) {
              this.setState({
                pending: false,
                error: message,
              });
            }.bind(this));
          }
        }.bind(this));
      } else {
        this.setState({
          pending: false,
          success: false,
        });
      }
    }.bind(this)).catch(function (err: any) {
      this.setState({
        pending: false,
        success: false,
        error: "Something went wrong. Please try again.",
      });
    }.bind(this));
  }

  validate(): Promise<boolean> {
    return Promise.all([
      this.name.validate(),
      this.email.validate(),
      this.password.validate(),
      this.recaptcha.validate()
    ]).then(function (values: Array<boolean>) {
      return values.every(v => { return v == true });
    });
  }

  open() {
    this.setState({
      open: true,
    });
  }

  close() {
    this.setState({
      open: false,
    });
  }

  render() {
    return (
      <Segment basic vertical>
        <Image size="medium" src="assets/botbox-masthead.svg" centered />
        <Container>
          <Grid padded centered>
            <Grid.Column largeScreen={5} computer={8} tablet={10} mobile={14}>
              <Segment>
                {(this.state.success) ? (
                  <Message success>
                    You have been successfully registered! Please check your email to finish verifying your account.
                </Message>
                ) : (
                    <Form error={this.state.error != undefined} loading={this.state.pending}>
                      {this.state.error != undefined ? <Message error>{this.state.error}</Message> : null}
                      <Name ref={e => this.name = e} />
                      <Email ref={e => this.email = e} />
                      <Password ref={e => this.password = e} />
                      <Recaptcha ref={e => this.recaptcha = e} />
                    </Form>
                  )}
                <Divider hidden />
                <Button primary disabled={this.state.pending} onClick={() => this.doRegister()}>Join</Button>
              </Segment>
            </Grid.Column>
          </Grid>
        </Container>
      </Segment>
    )
  }
}


