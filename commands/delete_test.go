package commands_test

import (
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	io_helpers "code.cloudfoundry.org/cli/testhelpers/io"
	. "github.com/cloudfoundry/v3-cli-plugin/commands"
	fakes "github.com/cloudfoundry/v3-cli-plugin/fakes"
)

var _ = Describe("Delete", func() {
	var (
		fakeCliConnection *fakes.FakeCliConnection
		args              = []string{"v3-delete", "my-app"}

		searchResult = `{
				"resources": [
					{
						"guid": "feed-dead-beef",
						"name": "my-app"
					}
				]
			}`
		searchResultWithSpace = `{
				"resources": [
					{
						"guid": "feed-dead-beef-with-space",
						"name": "my app"
					}
				]
			}`
		searchError error

		deleteResult = `{
			}`
		deleteError error
	)

	JustBeforeEach(func() {
		fakeCliConnection = &fakes.FakeCliConnection{}
		fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
			if reflect.DeepEqual(args, []string{"curl", "/v3/apps?names=my-app"}) {
				return []string{searchResult}, searchError
			} else if reflect.DeepEqual(args, []string{"curl", "/v3/apps?names=my+app"}) {
				return []string{searchResultWithSpace}, searchError
			} else if reflect.DeepEqual(args, []string{"curl", "/v3/apps/feed-dead-beef", "-X", "DELETE"}) {
				return []string{deleteResult}, deleteError
			} else if reflect.DeepEqual(args, []string{"curl", "/v3/apps/feed-dead-beef-with-space", "-X", "DELETE"}) {
				return []string{deleteResult}, deleteError
			}
			return []string{""}, nil
		}
	})

	It("deletes the app", func() {
		output := io_helpers.CaptureOutput(func() { Delete(fakeCliConnection, args) })
		Expect(fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(1)).
			To(Equal([]string{"curl", "/v3/apps/feed-dead-beef", "-X", "DELETE"}))

		Expect(output[0]).To(Equal("Deleting app my-app..."))
		Expect(output[1]).To(Equal("OK"))
	})

	Context("When the app name has a space", func() {
		It("deletes the app", func() {
			commandArgsWithSpace := []string{"v3-delete", "my app"}

			output := io_helpers.CaptureOutput(func() { Delete(fakeCliConnection, commandArgsWithSpace) })
			Expect(fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(1)).
				To(Equal([]string{"curl", "/v3/apps/feed-dead-beef-with-space", "-X", "DELETE"}))

			Expect(output[0]).To(Equal("Deleting app my app..."))
			Expect(output[1]).To(Equal("OK"))
		})
	})

	Context("The app fails to delete", func() {
		BeforeEach(func() {
			deleteError = errors.New("")
		})

		It("says that the delete failed", func() {
			output := io_helpers.CaptureOutput(func() { Delete(fakeCliConnection, args) })
			Expect(fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(1)).
				To(Equal([]string{"curl", "/v3/apps/feed-dead-beef", "-X", "DELETE"}))

			Expect(output[0]).To(Equal("Deleting app my-app..."))
			Expect(output[1]).To(Equal("Failed to delete app my-app"))
		})
	})

	Context("The app doesn't exist", func() {
		BeforeEach(func() {
			searchResult = `{
				"resources": [
				]
			}`
		})

		It("tells you the app wasn't found", func() {
			output := io_helpers.CaptureOutput(func() { Delete(fakeCliConnection, args) })
			Expect(fakeCliConnection.CliCommandWithoutTerminalOutputCallCount()).
				To(Equal(1))

			Expect(output[0]).To(Equal("Deleting app my-app..."))
			Expect(output[1]).To(Equal("App my-app not found"))
		})
	})
})
