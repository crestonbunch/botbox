import * as React from "react";
import "whatwg-fetch"
import { Grid, Divider, Popup, Label, Message, Input, Icon, Form, Modal, Button } from "semantic-ui-react";
import { Recaptcha } from "./recaptcha";
import { Username } from "./username";
import { Email } from "./email";
import { Password } from "./password";
import { Github } from "./github";

const PASSWORD_LENGTH = 6;

export interface RegisterProps {
  trigger: JSX.Element;
}

export interface RegisterState {
  open?: boolean;
  pending?: boolean;
  success?: boolean;
  error?: string;
}

export class Register extends React.Component<RegisterProps, RegisterState> {

  private username: Username;
  private email: Email;
  private password: Password;
  private recaptcha: Recaptcha;
  private github: Github;

  constructor(props: RegisterProps) {
    super(props);
    this.state = {
      open: false,
      pending: false,
      success: false
    };
  }

  doRegister(e: Event) {
    e.preventDefault();

    this.setState({
      pending: true,
      success: false,
      error: undefined,
    });

    this.validate().then(function(valid: boolean) {
      if (valid) {

        let request = {
          username: this.username.getValue(),
          email: this.email.getValue(),
          password: this.password.getValue(),
          captcha: this.recaptcha.getValue(),
        }

        fetch('/api/account/botbox/new', {
          method: 'POST',
          body: JSON.stringify(request)
        }).then(function(response: Response) {
          this.recaptcha.reset();
          if (response.status == 200) {
            this.setState({
              pending: false,
              success: true,
            });
          } else {
            response.text().then(function(message: string) {
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
    }.bind(this)).catch(function(err: any) {
      this.setState({
        pending: false,
        success: false,
      });
    }.bind(this));
  }

  validate(): Promise<boolean> {
    return Promise.all([
      this.username.validate(),
      this.email.validate(),
      this.password.validate(),
      this.recaptcha.validate()
    ]).then(function(values: Array<boolean>) {
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
    /*
    const form = (
      <Popup.Content>
        {(this.state.success) ? (
          <Message success>
            You have been successfully registered! Please check your email to finish verifying your account.
          </Message>
        ) : (
          <Form error={this.state.error != undefined} loading={this.state.pending}>
            {this.state.error != undefined ? <Message error>{this.state.error}</Message> : null}
            <Username ref={e => this.username = e} />
            <Email ref={e => this.email = e} />
            <Password ref={e => this.password = e} />
            <Recaptcha ref={e => this.recaptcha = e} />
            <Button positive disabled={this.state.pending} onClick={this.doRegister.bind(this)}>Join</Button>
          </Form>
        )}
      </Popup.Content>
    )

    return (<Popup
      trigger={this.props.trigger}
      content={form}
      on='focus'
      positioning='bottom center'
      wide='very'
    />)
     */

    const modal = (<Modal size="small" trigger={this.props.trigger} open={this.state.open} onOpen={this.open.bind(this)}
      onClose={this.close.bind(this)}>
      <Modal.Header><Icon name="add user" /> Register</Modal.Header>
      <Modal.Content>
        {(this.state.success) ? (
          <Message success>
            You have been successfully registered! Please check your email to finish verifying your account.
          </Message>
        ) : (
          <div style={{position: 'relative'}}>
            <Grid columns="two" centered>
              <Grid.Column>
                <Form error={this.state.error != undefined} loading={this.state.pending}>
                  {this.state.error != undefined ? <Message error>{this.state.error}</Message> : null}
                  <Username ref={e => this.username = e} />
                  <Email ref={e => this.email = e} />
                  <Password ref={e => this.password = e} />
                  <Recaptcha ref={e => this.recaptcha = e} />
                </Form>
              </Grid.Column>
              <Divider vertical>or</Divider>
              <Grid.Column verticalAlign="middle">
                <Github ref={e => this.github = e} />
              </Grid.Column>
            </Grid>
          </div>
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

