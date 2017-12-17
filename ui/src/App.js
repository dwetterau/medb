import React, { Component } from 'react';
import PagedrawGeneratedPage from './pagedraw/app'

class App extends Component {
  render() {
    return (
        <PagedrawGeneratedPage
          rootFolderList={[
              {
                  name: "folder1",
                  state: "collapsed",
              },
              {
                  name: "file1",
                  state: "file",
              },
              {
                  name: "folder2Empty",
                  state: "expandedEmpty",
              },
              {
                  name: "folder3",
                  state: "expanded",
                  contents: [
                      {
                          name: "subfile",
                          state: "file"
                      },
                      {
                          name: "subfile2",
                          state: "file"
                      },
                      {
                          name: "folder4",
                          state: "expanded",
                          contents: [
                              {
                                  name: "subsubfile1",
                                  state: "file",
                              }
                          ]
                      }
                  ]
              },
          ]}
        />
    );
  }
}

export default App;
