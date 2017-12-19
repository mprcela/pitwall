package cmd

import (
	"fmt"

	"github.com/minus5/pitwall/monit"
	"github.com/spf13/cobra"
)

// grepCmd represents the grep command
var grepCmd = &cobra.Command{
	Use:   "grep",
	Short: "Grep logs from the AWS CloudWatch",
	Long: fmt.Sprintf(`Grep logs from the AWS CloudWatch.
Filter patern refernece: https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/FilterAndPatternSyntax.html

  Supported time patterns (for start and end time):%s

  Examples:
    monit grep haproxy -s 16:23 -e 16:24
    monit grep haproxy -s "14.04. 16:23" -e "14.04. 16:24" --dry
    monit grep haproxy -s "3 hours ago"
    monit grep haproxy -d pg1 -s "3 hours ago"
    monit grep backend_api -e "15.04. 13:02" -s "15.04. 12:58" -f '"49B69912-5537-4553-48EB-99500B9FC539"' -p
    monit grep backend_api -s "1 day ago" -f "{$.retry>100}"`, monit.TimePatterns()),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			cmd.Usage()
			return
		}
		service := ""
		if len(args) == 1 {
			service = args[0]
		}

		st, errSt := monit.ParseTime(startTime)
		et, errEt := monit.ParseTime(endTime)
		if errSt != nil || errEt != nil {
			cmd.Usage()
			return
		}

		monit.Grep(monit.GrepOptions{
			Address:   getServiceAddress("nsq-to-cloudwatch", "nsq_to_cloudwatch"),
			Service:   service,
			Json:      json,
			Pretty:    pretty,
			Exclude:   splitComma(exclude),
			Include:   splitComma(include),
			Filter:    filter,
			StartTime: st,
			EndTime:   et,
		})
	},
}

var (
	filter    string
	startTime string
	endTime   string
)

func init() {
	rootCmd.AddCommand(grepCmd)

	grepCmd.Flags().StringVarP(&dc, "dc", "d", "", "datacenter to use")
	grepCmd.MarkFlagRequired("dc")

	grepCmd.Flags().BoolVarP(&json, "json", "j", false, "print unparsed json log line")
	grepCmd.Flags().BoolVarP(&pretty, "pretty", "p", false, "pretrty print json log line")
	grepCmd.Flags().StringVarP(&exclude, "exclude", "x", "", "list of attributes to EXCLUDE separated by ,")
	grepCmd.Flags().StringVarP(&include, "include", "i", "", "list of attributes to INCLUDE separated by ,")

	grepCmd.Flags().StringVarP(&filter, "filter", "f", "", "AWS CloudWatch filter pattern (see ref.)")

	grepCmd.Flags().StringVarP(&startTime, "start-time", "s", "", "find logs from start_time")
	grepCmd.Flags().StringVarP(&endTime, "end-time", "e", "", "find logs till end_time")
}
