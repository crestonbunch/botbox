import * as React from "react";
import "whatwg-fetch"
import { Label, Message, Input, Icon, Form, Button } from "semantic-ui-react";

const USERNAME_LENGTH = 20;

export interface UsernameProps {
}

export interface UsernameState {
  error?: string
  value?: string
}

export class Username extends React.Component<UsernameProps, UsernameState> {

  constructor(props: UsernameProps) {
    super(props);
    this.state = {
      value: "",
    };
  }

  getValue(): string {
    return this.state.value;
  }

  updateUsername(e: Event) {
    let value = (e.target as any).value;

    this.setState({
      value: value,
      error: undefined
    }, this.validate);
  }

  validate(): Promise<boolean> {
    if (this.state.value == "") {
      this.setState({
        error: "Please provide a username."
      });
      return new Promise<boolean>(function(r) {r(false);});
    } else if (this.state.value.length > USERNAME_LENGTH) {
      this.setState({
        error: "Username must be no more than 20 characters."
      });
      return new Promise<boolean>(function(r) {r(false);});
    }

    if (this.state.value.length > 0) {
      return fetch('/api/account/exists/username/' + this.state.value).
        then(function(response: Response) {
          if (response.status == 200) {
            return response.text();
          } else {
            throw "Invalid username!"
          }
      }.bind(this)).catch(function(err: any) {
        return false;
      }.bind(this)).then(function(result: string) {
        if (result == "true") {
          this.setState({
            error: "That username is already taken!"
          });

          return false;
        } else {
          return true;
        }
      }.bind(this)).catch(function(err: any) {
        return false;
      }.bind(this));
    }

    return new Promise<boolean>(function(r) {r(true);});
  }

  render() {
    return (
    <Form.Field required error={this.state.error != undefined}>
      <label>Username</label>
      {this.state.error != undefined ?
        <Label basic pointing="below" color="red"><Icon name="warning sign" />
          {this.state.error}
        </Label>
      : null}
      <Input icon='users' iconPosition='left'
        onChange={this.updateUsername.bind(this)}
        name="username" placeholder="Enter a unique username" />
    </Form.Field>
    )
  }

}
