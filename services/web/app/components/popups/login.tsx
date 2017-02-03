import * as React from "react";
import "whatwg-fetch";
import { observer } from "mobx-react";
import { observable, autorun, action } from "mobx";
import { Message, Input, Popup, Button, Form } from "semantic-ui-react";
import { Api } from '../../api';
import { Store } from "../../store";
import { SessionData } from "../../models";

export interface LoginPopupProps {
  trigger: JSX.Element;
  store: Store;
}

@observer
export class LoginPopup extends React.Component<LoginPopupProps, {}> {

  @observable email: string = "";
  @observable password: string = "";

  @observable error: string = "";

  // tracks if an asynchronous operation is running
  // that should disable the form until it is done.
  @observable busy: boolean = false;

  @action doLogin() {
    this.busy = true;
    Api.login(this.email, this.password).then((session: SessionData) => {
      this.props.store.login(session);
      this.busy = false;
    }).catch((reason: any) => {
      this.error = reason as any;
      this.busy = false;
    });
  }

  render() {
    const store = this.props.store;

    const errMsg = (this.error !== "") ? (
      <Message error>{this.error}</Message>
    ) : null;

    const form = (<Form loading={this.busy}>
      {errMsg}
      <Form.Field>
        <Input icon='mail' iconPosition='left' placeholder="Email"
          onChange={(_, val) => this.email = val.value} />
      </Form.Field>
      <Form.Field>
        <Input type="password" icon='lock' iconPosition='left'
          placeholder="Password"
          onChange={(_, val) => this.password = val.value} />
      </Form.Field>
      <Button primary onClick={(e) => { e.preventDefault(); this.doLogin() }}>
        Login
      </Button>
    </Form>)

    return (<Popup
      trigger={this.props.trigger}
      content={form}
      on='click'
      positioning='bottom center'
    />)
  }

}