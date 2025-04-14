package reconcilers

import (
	"testing"
	"time"

	reconcilerutil "github.com/3scale-sre/basereconciler/util"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var t1, t2 time.Time

func init() {
	t1, _ = time.Parse(time.RFC3339, "2020-12-19T00:00:00Z")
	t2, _ = time.Parse(time.RFC3339, "2020-12-20T00:00:00Z")

}

func TestIsStatusReconciled(t *testing.T) {
	type args struct {
		dsc             *operatorv1alpha1.DiscoveryServiceCertificate
		certificateHash string
		ready           bool
		notBefore       time.Time
		notAfter        time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Returns true, status up to date",
			args: args{
				dsc: &operatorv1alpha1.DiscoveryServiceCertificate{
					Status: operatorv1alpha1.DiscoveryServiceCertificateStatus{
						Ready:           reconcilerutil.Pointer(true),
						CertificateHash: reconcilerutil.Pointer("xxxx"),
						NotBefore:       &metav1.Time{Time: t1},
						NotAfter:        &metav1.Time{Time: t2},
						Conditions:      []metav1.Condition{},
					},
				},
				certificateHash: "xxxx",
				ready:           true,
				notBefore:       t1,
				notAfter:        t2,
			},
			want: true,
		},
		{
			name: "Returns false, ready needs update",
			args: args{
				dsc: &operatorv1alpha1.DiscoveryServiceCertificate{
					Status: operatorv1alpha1.DiscoveryServiceCertificateStatus{
						Ready:           reconcilerutil.Pointer(false),
						CertificateHash: reconcilerutil.Pointer("xxxx"),
						NotBefore:       &metav1.Time{Time: t1},
						NotAfter:        &metav1.Time{Time: t2},
						Conditions:      []metav1.Condition{},
					},
				},
				certificateHash: "xxxx",
				ready:           true,
				notBefore:       t1,
				notAfter:        t2,
			},
			want: false,
		},
		{
			name: "Returns false, certificateHash needs update",
			args: args{
				dsc: &operatorv1alpha1.DiscoveryServiceCertificate{
					Status: operatorv1alpha1.DiscoveryServiceCertificateStatus{
						Ready:           reconcilerutil.Pointer(true),
						CertificateHash: reconcilerutil.Pointer("xxxx"),
						NotBefore:       &metav1.Time{Time: t1},
						NotAfter:        &metav1.Time{Time: t2},
						Conditions:      []metav1.Condition{},
					},
				},
				certificateHash: "zzzz",
				ready:           true,
				notBefore:       t1,
				notAfter:        t2,
			},
			want: false,
		},
		{
			name: "Returns false, NotBefore needs update",
			args: args{
				dsc: &operatorv1alpha1.DiscoveryServiceCertificate{
					Status: operatorv1alpha1.DiscoveryServiceCertificateStatus{
						Ready:           reconcilerutil.Pointer(true),
						CertificateHash: reconcilerutil.Pointer("xxxx"),
						NotBefore:       &metav1.Time{},
						NotAfter:        &metav1.Time{Time: t2},
						Conditions:      []metav1.Condition{},
					},
				},
				certificateHash: "zzzz",
				ready:           true,
				notBefore:       t1,
				notAfter:        t2,
			},
			want: false,
		},
		{
			name: "Returns false, NotAfter needs update",
			args: args{
				dsc: &operatorv1alpha1.DiscoveryServiceCertificate{
					Status: operatorv1alpha1.DiscoveryServiceCertificateStatus{
						Ready:           reconcilerutil.Pointer(true),
						CertificateHash: reconcilerutil.Pointer("xxxx"),
						NotBefore:       &metav1.Time{Time: t1},
						NotAfter:        &metav1.Time{},
						Conditions:      []metav1.Condition{},
					},
				},
				certificateHash: "zzzz",
				ready:           true,
				notBefore:       t1,
				notAfter:        t2,
			},
			want: false,
		},
		{
			name: "Returns false, conditions need init",
			args: args{
				dsc: &operatorv1alpha1.DiscoveryServiceCertificate{
					Status: operatorv1alpha1.DiscoveryServiceCertificateStatus{
						Ready:           reconcilerutil.Pointer(true),
						CertificateHash: reconcilerutil.Pointer("xxxx"),
						NotBefore:       &metav1.Time{Time: t1},
						NotAfter:        &metav1.Time{Time: t1},
					},
				},
				certificateHash: "zzzz",
				ready:           true,
				notBefore:       t1,
				notAfter:        t2,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStatusReconciled(tt.args.dsc, tt.args.certificateHash,
				tt.args.ready, tt.args.notBefore, tt.args.notAfter); got != tt.want {
				t.Errorf("IsStatusReconciled() = %v, want %v", got, tt.want)
			}
		})
	}
}
