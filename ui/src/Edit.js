import React, { Component } from 'react';
import Notifications, {notify} from 'react-notify-toast';
import PagedrawGeneratedPage from './pagedraw/editfile'
import $ from 'jquery';
import {emptyGitInfo, fetchGitInfo, handlePull, handlePush} from './Api'

class EditFile extends Component {
  constructor(props) {
      super();
      this.state = {
          fileContent: props.file.content,
          viewState: "viewing",
          gitInfo: emptyGitInfo(),
      };

      // Kick off populating git info
      this.populateGitInfo();
  }

  componentWillReceiveProps(nextProps) {
      this.setState({
          fileContent: nextProps.file.content,
      });

      // Kick off populating git info
      this.populateGitInfo();
  }

  populateGitInfo() {
      fetchGitInfo((newInfo) => {
          this.setState({
              gitInfo: newInfo,
          })
      })
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
          this.populateGitInfo()
      })
  }

  handleViewStateChange() {
      let newViewState = (this.state.viewState === "viewing") ? "editing" : "viewing";
      this.setState({
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
              lastCommit={this.state.gitInfo.lastCommit}
              lastPull={this.state.gitInfo.lastPull}
              remoteAheadBy={this.state.gitInfo.remoteAheadBy}
              localAheadBy={this.state.gitInfo.localAheadBy}
          />
          <Notifications />
      </div>;
  }
}

export default EditFile;
