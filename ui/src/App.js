import React, { Component } from 'react';
import PagedrawGeneratedPage from './pagedraw/app'

class App extends Component {
  render() {
    return (
        <PagedrawGeneratedPage
          rootFolderList={this.props.rootFolderList}
        />
    );
  }
}

export default App;
