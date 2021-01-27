package experiment_test

import (
	"context"

	"github.com/iter8-tools/handler/experiment"
	"github.com/iter8-tools/handler/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Experiment's handler field", func() {
	Context("when containing handler actions", func() {
		var exp *experiment.Experiment
		var err error
		It("should retrieve handler info properly", func() {
			By("reading the experiment from file")
			exp, err = (&experiment.Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment1.yaml")).Build()
			Expect(err).ToNot(HaveOccurred())

			By("creating experiment in cluster")
			Expect(k8sClient.Create(context.Background(), exp)).To(Succeed())

			By("fetching experiment from cluster")
			b := &experiment.Builder{}
			exp2, err := b.FromCluster("sklearn-iris-experiment-1", "default", k8sClient).Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(exp2.Spec).To(Equal(exp.Spec))
		})

		It("should run handler", func() {
			By("reading the experiment from file")
			exp, err = (&experiment.Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment6.yaml")).Build()
			Expect(err).ToNot(HaveOccurred())

			By("creating experiment in cluster")
			Expect(k8sClient.Create(context.Background(), exp)).To(Succeed())
			Expect(k8sClient.Status().Update(context.Background(), exp)).To(Succeed())

			By("fetching experiment from cluster")
			b := &experiment.Builder{}
			exp2, err := b.FromCluster("sklearn-iris-experiment-6", "default", k8sClient).Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(exp2.Spec).To(Equal(exp.Spec))

			By("running the experiment")
			err = exp2.Run("start")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should deal with extrapolation errors", func() {
			By("reading the experiment from file")
			exp, err = (&experiment.Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment7.yaml")).Build()
			Expect(err).ToNot(HaveOccurred())

			By("running and gracefully exiting")
			err = exp.Run("start")
			Expect(err).To(HaveOccurred())
		})

	})
})
