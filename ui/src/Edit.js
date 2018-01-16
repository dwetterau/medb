import React, { Component } from 'react';
import Notifications, {notify} from 'react-notify-toast';
import PagedrawGeneratedPage from './pagedraw/editfile'
import $ from 'jquery';
import {handlePull, handlePush} from './Api'

class EditFile extends Component {
  constructor(props) {
      super();
      this.state = {
          fileContent: props.file.content,
          viewState: "viewing",
      }
  }

  handleContentChange(e) {
      this.setState({
          fileContent: e.target.value,
      })
  }

  handleCommit() {
      $.post("/api/1/edit", {
          fileID: this.props.file.id,
          fileContent: this.state.fileContent,
      }).done(() => {
          notify.show("Committed successfully.");
      })
  }

  handleViewStateChange() {
      let newViewState = (this.state.viewState === "viewing") ? "editing" : "viewing";
      this.setState({
          fileContent: this.state.fileContent,
          viewState: newViewState,
      })
  }

  render() {
      return <div>
          <PagedrawGeneratedPage
              fileID={this.props.file.id}
              fileName={this.props.file.name}
              fileContent={this.state.fileContent}
              handleContentChange={this.handleContentChange.bind(this)}
              handleCommit={this.handleCommit.bind(this)}
              handlePull={handlePull}
              handlePush={handlePush}
              handleViewStateChange={this.handleViewStateChange.bind(this)}
              viewState={this.state.viewState}
          />
          <Notifications />
      </div>;
  }
}

export default EditFile;
