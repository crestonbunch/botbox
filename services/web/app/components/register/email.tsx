import * as React from "react";
import "whatwg-fetch"
import { Label, Message, Input, Icon, Form, Button } from "semantic-ui-react";

export interface EmailProps {
}

export interface EmailState {
  error?: string
  value?: string
}

export class Email extends React.Component<EmailProps, EmailState> {

  constructor(props: EmailProps) {
    super(props);
    this.state = {
      value: "",
    };
  }

  getValue(): string {
    return this.state.value;
  }

  updateEmail(e: Event) {
    let value = (e.target as any).value;

    this.setState({
      value: value,
      error: undefined
    }, this.checkExists.bind(this));
  }

  checkExists(): Promise<boolean> {

    if (this.state.value.length > 0) {
      return fetch('/api/account/exists/email/' + this.state.value).
        then(function(response: Response) {
          if (response.status == 200) {
            return response.text();
          } else {
            throw "Invalid email!"
          }
      }.bind(this)).catch(function(err: any) {
        return false;
      }.bind(this)).then(function(result: string) {
        if (result == "true") {
          this.setState({
            error: "That email is already in use!"
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

  validate(): Promise<boolean> {
    if (this.state.value == "") {
      this.setState({
        error: "Please provide an email."
      });
      return new Promise<boolean>(function(r) {r(false);});
    } else if (this.state.value.indexOf('@') < 0) {
      this.setState({
        error: "Please provide a valid email."
      });
      return new Promise<boolean>(function(r) {r(false);});
    }

    return this.checkExists();
  }

  render() {
    return (
    <Form.Field required error={this.state.error != undefined}>
      <label>Email</label>
      {this.state.error != undefined ?
        <Label basic pointing="below" color="red"><Icon name="warning sign" />
          {this.state.error}
        </Label>
      : null}
      <Input icon='mail' iconPosition='left'
        onChange={this.updateEmail.bind(this)}
        name="email" placeholder="email@example.com" />
    </Form.Field>
    )
  }

}
