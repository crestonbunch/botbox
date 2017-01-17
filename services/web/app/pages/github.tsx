import * as React from "react";
import "whatwg-fetch"
import { Loader, Dimmer, Form, Message, Button } from "semantic-ui-react"
import { Username } from "../components/register/username"

export interface GithubLoginProps {
  location?: any
}

export interface GithubLoginState {
  step: "token" | "conflict" | "merge" | "error" | "success" | "pending",
  token?: string,
  username?: string,
  email?: string,
  error?: string
}

interface TokenResponse {
  username: string,
  email: string,
  token: string
  state: "merge" | "conflict" | "register" | "login"
}

export class GithubLogin extends React.Component<GithubLoginProps, GithubLoginState> {

  username: Username;

  constructor(props: GithubLoginProps) {
    super(props)
    this.setState({
      step: "token",
    });
  }

  componentDidMount() {
    this.getToken();
  }

  getToken() {

    this.setState({
      step: "pending",
    });

    let request = {
      state: this.props.location.query.state,
      code: this.props.location.query.code,
    }
    fetch('/api/session/github/login', {
      method: 'POST',
      body: JSON.stringify(request)
    }).then(function(response: Response) {
      if (response.status == 200) {
        response.json().then(function(message: TokenResponse) {
          switch (message.state) {
            case "merge":
              this.setState({
                username: message.username,
                email: message.email,
                token: message.token,
                step: "merge"
              });
              break;
            case "conflict":
              this.setState({
                username: message.username,
                email: message.email,
                token: message.token,
                step: "conflict"
              });
              break;
            case "register":
              this.setState({
                step: "success"
              });
              break;
            case "login":
              this.setState({
                step: "success"
              });
              break;
            default:
              this.setState({
                error: "An unknown error occurred.",
                step: "error"
              });
              break;
          }
        });
      } else {
        response.text().then(function(message: string) {
          this.setState({
            step: "error",
            error: message,
          });
        }.bind(this));
      }
    }.bind(this));
  }

  doRegister() {
    this.setState({
      step: "pending"
    });

    this.validate().then(function(valid: boolean) {
      if (valid) {

        let request = {
          username: this.username.getValue(),
          email: this.email.getValue(),
          token: this.recaptcha.getValue(),
        }

        fetch('/api/account/github/new', {
          method: 'POST',
          body: JSON.stringify(request)
        }).then(function(response: Response) {
          if (response.status == 200) {
            this.setState({
              step: "success",
            });
          } else {
            response.text().then(function(message: string) {
              this.setState({
                step: "error",
                error: message,
              });
            }.bind(this));
          }
        }.bind(this));
      } else {
        this.setState({
          step: "error",
          error: "Invalid registration attempt."
        });
      }
    }.bind(this)).catch(function(err: any) {
      this.setState({
        step: "error",
        error: "An unknown error occurred."
      });
    }.bind(this));
  }

  validate(): Promise<boolean> {
    if (this.username != undefined) {
      return this.username.validate();
    } else {
      return new Promise<boolean>((r) => r(true));
    }
  }

  fixConflict() {
    this.setState({
      step: "pending",
    },() =>
      this.validate().then(function(result: boolean) {
        if (result == false) {
          this.setState({
            step: "conflict",
          });
        } else {
          this.setState({
            step: "register",
            username: this.username.getValue(),
          }, this.doRegister)
        }
      }.bind(this))
    );
  }

  doMerge() {
    this.doRegister();
  }

  cancel() {
    window.close();
  }

  render() {
    switch (this.state.step) {
      case "pending":
        return (
          <Dimmer active><Loader /></Dimmer>
      )
      case "success":
        return (
          <Message success>
            Login successful! You may close this window.
          </Message>
        );
      case "error":
        return (
          <Message error>
            {this.state.error}
          </Message>
        );
      case "conflict":
        return (
          <Form error={this.state.error != undefined}>
            The username on your GitHub account {this.state.username} is already being used by another
            user. Please provide a different name you want to be recognized as on Botbox. 
            {this.state.error != undefined ? <Message error>{this.state.error}</Message> : null}
            <Username ref={e => this.username = e} />
            <Button primary onClick={this.fixConflict}>Finish</Button>
          </Form>
        );
      case "merge":
      return (
        <div>
          An email associated with your GitHub account is already registered! Would you
          like to merge the accounts?
          <Button onClick={this.cancel}>No</Button>
          <Button primary onClick={this.doMerge}>Yes</Button>
        </div>
      );
    }
  }
}

