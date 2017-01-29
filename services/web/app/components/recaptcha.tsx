import * as React from "react";
import 'grecaptcha'
import { observer } from "mobx-react";
import { Label, Message, Input, Icon, Form, Button } from "semantic-ui-react";

const RECAPTCHA_ID = "recaptcha";

export interface RecaptchaProps {
  error: string,
  sitekey: string,
  onChange: (val: string) => void,
}

export interface RecaptchaState {
  ready?: boolean,
  readyCheck?: any,
}

@observer
export class Recaptcha extends React.Component<RecaptchaProps, RecaptchaState> {

  constructor(props: RecaptchaProps) {
    super(props);
    this.state = {
      ready: false,
      readyCheck: setInterval(this.checkReady.bind(this), 200),
    }
  }

  private renderRecaptcha() {
    if (this.state.ready) {
      this.forceUpdate();
      grecaptcha.render(
        RECAPTCHA_ID,
        {
          sitekey: this.props.sitekey,
          callback: this.props.onChange,
        }
      );
    }
  }

  private checkReady() {
    if (grecaptcha != undefined) {
      this.setState({
        ready: true,
      });
      this.renderRecaptcha();
      clearInterval(this.state.readyCheck);
    }
  }

  reset() {
    if (grecaptcha != undefined) {
      grecaptcha.reset();
    }
  }

  render() {
    if (!this.state.ready) {
      return <Icon size="large" name="notched circle" loading />
    }
    return (
      <Form.Field required error={this.props.error != ""}>
        <label>Singularity check</label>
        {this.props.error != "" ? <Label basic pointing="below" color="red">{this.props.error}</Label> : null}
        <div id={RECAPTCHA_ID} data-sitekey={this.props.sitekey}></div>
      </Form.Field>
    );
  }
}

